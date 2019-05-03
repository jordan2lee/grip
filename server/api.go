package server

import (
	"fmt"
	"io"
	"time"

	"github.com/bmeg/grip/engine"
	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Traversal parses a GraphQuery request and streams the results back
func (server *GripServer) Traversal(query *gripql.GraphQuery, queryServer gripql.Query_TraversalServer) error {
	start := time.Now()
	resultsChan, err := engine.Traversal(queryServer.Context(), server.db, server.conf.WorkDir, query)
	if err != nil {
		return err
	}
	for row := range resultsChan {
		if err == nil {
			err = queryServer.Send(row)
		}
	}
	if err != nil {
		return fmt.Errorf("error sending Traversal result: %v", err)
	}
	log.WithFields(log.Fields{"query": query, "elapsed_time": time.Since(start)}).Debug("Traversal")
	return nil
}

// ListGraphs returns a list of graphs managed by the driver
func (server *GripServer) ListGraphs(ctx context.Context, empty *gripql.Empty) (*gripql.ListGraphsResponse, error) {
	graphs := server.db.ListGraphs()
	return &gripql.ListGraphsResponse{Graphs: graphs}, nil
}

// GetVertex returns a vertex given a gripql.Element
func (server *GripServer) GetVertex(ctx context.Context, elem *gripql.ElementID) (*gripql.Vertex, error) {
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}
	o := graph.GetVertex(elem.Id, true)
	if o == nil {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("vertex %s not found", elem.Id))
	}
	return o, nil
}

// GetEdge returns an edge given a gripql.Element
func (server *GripServer) GetEdge(ctx context.Context, elem *gripql.ElementID) (*gripql.Edge, error) {
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}
	o := graph.GetEdge(elem.Id, true)
	if o == nil {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("edge %s not found", elem.Id))
	}
	return o, nil
}

// GetTimestamp returns the update timestamp of a graph
func (server *GripServer) GetTimestamp(ctx context.Context, elem *gripql.GraphID) (*gripql.Timestamp, error) {
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}
	return &gripql.Timestamp{Timestamp: graph.GetTimestamp()}, nil
}

// DeleteGraph deletes a graph
func (server *GripServer) DeleteGraph(ctx context.Context, elem *gripql.GraphID) (*gripql.EditResult, error) {
	err := server.db.DeleteGraph(elem.Graph)
	if err != nil {
		return nil, fmt.Errorf("DeleteGraph: deleting graph %s: %v", elem.Graph, err)
	}
	schemaName := fmt.Sprintf("%s%s", elem.Graph, gripql.SchemaSuffix)
	if engine.GraphExists(server.db, schemaName) {
		err := server.db.DeleteGraph(schemaName)
		if err != nil {
			return nil, fmt.Errorf("DeleteGraph: deleting schema for graph %s: %v", elem.Graph, err)
		}
	}
	return &gripql.EditResult{Id: elem.Graph}, nil
}

// AddGraph creates a new graph on the server
func (server *GripServer) AddGraph(ctx context.Context, elem *gripql.GraphID) (*gripql.EditResult, error) {
	err := gripql.ValidateGraphName(elem.Graph)
	if err != nil {
		return nil, err
	}
	err = server.db.AddGraph(elem.Graph)
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: elem.Graph}, err
}

// AddVertex adds a vertex to the graph
func (server *GripServer) AddVertex(ctx context.Context, elem *gripql.GraphElement) (*gripql.EditResult, error) {
	if gripql.IsSchema(elem.Graph) {
		return nil, fmt.Errorf("unable to add vertex to graph schema; use AddSchema")
	}
	return server.addVertex(ctx, elem)
}

func (server *GripServer) addVertex(ctx context.Context, elem *gripql.GraphElement) (*gripql.EditResult, error) {
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}

	vertex := elem.Vertex
	err = vertex.Validate()
	if err != nil {
		return nil, fmt.Errorf("vertex validation failed: %v", err)
	}

	err = graph.AddVertex([]*gripql.Vertex{vertex})
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: elem.Vertex.Gid}, nil
}

// AddEdge adds an edge to the graph
func (server *GripServer) AddEdge(ctx context.Context, elem *gripql.GraphElement) (*gripql.EditResult, error) {
	if gripql.IsSchema(elem.Graph) {
		return nil, fmt.Errorf("unable to add edge to graph schema; use AddSchema")
	}
	return server.addEdge(ctx, elem)
}

func (server *GripServer) addEdge(ctx context.Context, elem *gripql.GraphElement) (*gripql.EditResult, error) {
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}

	edge := elem.Edge
	if edge.Gid == "" {
		edge.Gid = util.UUID()
	}
	err = edge.Validate()
	if err != nil {
		return nil, fmt.Errorf("edge validation failed: %v", err)
	}

	err = graph.AddEdge([]*gripql.Edge{edge})
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: edge.Gid}, nil
}

type graphElementArray struct {
	graph    string
	vertices []*gripql.Vertex
	edges    []*gripql.Edge
}

func newGraphElementArray(name string, vertexBufSize, edgeBufSize int) *graphElementArray {
	if vertexBufSize != 0 {
		return &graphElementArray{graph: name, vertices: make([]*gripql.Vertex, 0, vertexBufSize)}
	}
	if edgeBufSize != 0 {
		return &graphElementArray{graph: name, edges: make([]*gripql.Edge, 0, edgeBufSize)}
	}
	return nil
}

// BulkAdd a stream of inputs and loads them into the graph
func (server *GripServer) BulkAdd(stream gripql.Edit_BulkAddServer) error {
	vertexBatchSize := 50
	edgeBatchSize := 50

	vertCount := 0
	edgeCount := 0

	vertexBatchChan := make(chan *graphElementArray)
	edgeBatchChan := make(chan *graphElementArray)
	closeChan := make(chan bool)

	go func() {
		for vBatch := range vertexBatchChan {
			if gripql.IsSchema(vBatch.graph) {
				err := "cannot add vertex to schema graph"
				log.WithFields(log.Fields{"error": err}).Error("BulkAdd: add vertex error")
				continue
			}
			if len(vBatch.vertices) > 0 && vBatch.graph != "" {
				graph, err := server.db.Graph(vBatch.graph)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("BulkAdd: graph connection error")
					continue
				}
				err = graph.AddVertex(vBatch.vertices)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("BulkAdd: add vertex error")
				}
			}
		}
		closeChan <- true
	}()

	go func() {
		for eBatch := range edgeBatchChan {
			if gripql.IsSchema(eBatch.graph) {
				err := "cannot add edge to schema graph"
				log.WithFields(log.Fields{"error": err}).Error("BulkAdd: add edge error")
				continue
			}
			if len(eBatch.edges) > 0 && eBatch.graph != "" {
				graph, err := server.db.Graph(eBatch.graph)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("BulkAdd: graph connection error")
					continue
				}
				err = graph.AddEdge(eBatch.edges)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Error("BulkAdd: add edge error")
				}
			}
		}
		closeChan <- true
	}()

	vertexBatch := newGraphElementArray("", vertexBatchSize, 0)
	edgeBatch := newGraphElementArray("", 0, edgeBatchSize)
	var loopErr error
	for loopErr == nil {
		element, err := stream.Recv()
		if err == io.EOF {
			if vertCount != 0 {
				log.Debugf("%d vertices streamed", vertCount)
			}
			if edgeCount != 0 {
				log.Debugf("%d edges streamed", edgeCount)
			}
			vertexBatchChan <- vertexBatch
			edgeBatchChan <- edgeBatch
			loopErr = err
		} else if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("BulkAdd: streaming error")
			loopErr = err
		} else {
			if element.Vertex != nil {
				if vertexBatch.graph != element.Graph || len(vertexBatch.vertices) >= vertexBatchSize {
					vertexBatchChan <- vertexBatch
					vertexBatch = newGraphElementArray(element.Graph, vertexBatchSize, 0)
				}
				vertex := element.Vertex
				err := vertex.Validate()
				if err != nil {
					return fmt.Errorf("vertex validation failed: %v", err)
				}
				vertexBatch.vertices = append(vertexBatch.vertices, vertex)
				vertCount++
			} else if element.Edge != nil {
				if edgeBatch.graph != element.Graph || len(edgeBatch.edges) >= edgeBatchSize {
					edgeBatchChan <- edgeBatch
					edgeBatch = newGraphElementArray(element.Graph, 0, edgeBatchSize)
				}
				edge := element.Edge
				if edge.Gid == "" {
					edge.Gid = util.UUID()
				}
				err := edge.Validate()
				if err != nil {
					return fmt.Errorf("edge validation failed: %v", err)
				}
				edgeBatch.edges = append(edgeBatch.edges, edge)
				edgeCount++
			}
		}
	}

	close(edgeBatchChan)
	close(vertexBatchChan)
	<-closeChan
	<-closeChan

	if loopErr != io.EOF {
		return loopErr
	}
	return stream.SendAndClose(&gripql.EditResult{})
}

// DeleteVertex deletes a vertex from the server
func (server *GripServer) DeleteVertex(ctx context.Context, elem *gripql.ElementID) (*gripql.EditResult, error) {
	if gripql.IsSchema(elem.Graph) {
		return nil, fmt.Errorf("unable to delete vertex from graph schema; use AddSchema")
	}
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}
	err = graph.DelVertex(elem.Id)
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: elem.Id}, nil
}

// DeleteEdge deletes an edge from the graph server
func (server *GripServer) DeleteEdge(ctx context.Context, elem *gripql.ElementID) (*gripql.EditResult, error) {
	if gripql.IsSchema(elem.Graph) {
		return nil, fmt.Errorf("unable to delete edge from graph schema; use AddSchema")
	}
	graph, err := server.db.Graph(elem.Graph)
	if err != nil {
		return nil, err
	}
	err = graph.DelEdge(elem.Id)
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: elem.Id}, nil
}

// AddIndex adds a new index
func (server *GripServer) AddIndex(ctx context.Context, idx *gripql.IndexID) (*gripql.EditResult, error) {
	if gripql.IsSchema(idx.Graph) {
		return nil, fmt.Errorf("unupported operation for graph schema")
	}
	graph, err := server.db.Graph(idx.Graph)
	if err != nil {
		return nil, err
	}
	err = graph.AddVertexIndex(idx.Label, idx.Field)
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: idx.Field}, nil
}

// DeleteIndex removes an index from the server
func (server *GripServer) DeleteIndex(ctx context.Context, idx *gripql.IndexID) (*gripql.EditResult, error) {
	if gripql.IsSchema(idx.Graph) {
		return nil, fmt.Errorf("unupported operation for graph schema")
	}
	graph, err := server.db.Graph(idx.Graph)
	if err != nil {
		return nil, err
	}
	err = graph.DeleteVertexIndex(idx.Label, idx.Field)
	if err != nil {
		return nil, err
	}
	return &gripql.EditResult{Id: idx.Field}, nil
}

// ListIndices lists avalible indices from a graph
func (server *GripServer) ListIndices(ctx context.Context, idx *gripql.GraphID) (*gripql.ListIndicesResponse, error) {
	graph, err := server.db.Graph(idx.Graph)
	if err != nil {
		return nil, err
	}
	indices := []*gripql.IndexID{}
	for i := range graph.GetVertexIndexList() {
		indices = append(indices, i)
	}
	return &gripql.ListIndicesResponse{Indices: indices}, nil
}

// ListLabels lists the vertex and edge labels in a graph
func (server *GripServer) ListLabels(ctx context.Context, idx *gripql.GraphID) (*gripql.ListLabelsResponse, error) {
	graph, err := server.db.Graph(idx.Graph)
	if err != nil {
		return nil, err
	}
	vLabels, err := graph.ListVertexLabels()
	if err != nil {
		return nil, err
	}
	eLabels, err := graph.ListEdgeLabels()
	if err != nil {
		return nil, err
	}
	return &gripql.ListLabelsResponse{VertexLabels: vLabels, EdgeLabels: eLabels}, nil
}

func (server *GripServer) buildSchemas(ctx context.Context) {
	for _, name := range server.db.ListGraphs() {
		select {
		case <-ctx.Done():
			return

		default:
			if gripql.IsSchema(name) {
				continue
			}
			log.WithFields(log.Fields{"graph": name}).Debug("building graph schema")
			schema, err := server.db.BuildSchema(ctx, name, server.conf.SchemaInspectN, server.conf.SchemaRandomSample)
			if err == nil {
				log.WithFields(log.Fields{"graph": name}).Debug("cached graph schema")
				err := server.addSchemaGraph(ctx, schema)
				if err != nil {
					log.WithFields(log.Fields{"graph": name, "error": err}).Error("failed to store graph schema")
				}
				server.schemas[name] = schema
			} else {
				log.WithFields(log.Fields{"graph": name, "error": err}).Error("failed to build graph schema")
			}
		}
	}
}

// cacheSchemas calls GetSchema on each graph and caches the schemas in memory
func (server *GripServer) cacheSchemas(ctx context.Context) {
	if server.db == nil {
		return
	}

	if time.Duration(server.conf.SchemaRefreshInterval) == 0 {
		server.buildSchemas(ctx)
		return
	}

	ticker := time.NewTicker(time.Duration(server.conf.SchemaRefreshInterval))
	server.buildSchemas(ctx)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			server.buildSchemas(ctx)
		}
	}
}

// GetSchema returns the schema of a specific graph in the database
func (server *GripServer) GetSchema(ctx context.Context, elem *gripql.GraphID) (*gripql.Graph, error) {
	if !engine.GraphExists(server.db, elem.Graph) {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("graph %s: not found", elem.Graph))
	}
	schema, ok := server.schemas[elem.Graph]
	if !ok {
		if server.conf.AutoBuildSchemas {
			return nil, grpc.Errorf(codes.Unavailable, fmt.Sprintf("graph %s: schema not available; try again later", elem.Graph))
		}
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("graph %s: schema not found", elem.Graph))
	}

	if schema.Graph == "" {
		schema.Graph = elem.Graph
	}

	return schema, nil
}

// AddSchema caches a graph schema on the server
func (server *GripServer) AddSchema(ctx context.Context, req *gripql.Graph) (*gripql.EditResult, error) {
	err := server.addSchemaGraph(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to store new schema: %v", err)
	}
	server.schemas[req.Graph] = req
	return &gripql.EditResult{}, err
}

func (server *GripServer) addSchemaGraph(ctx context.Context, schema *gripql.Graph) error {
	schemaName := fmt.Sprintf("%s%s", schema.Graph, gripql.SchemaSuffix)
	if engine.GraphExists(server.db, schemaName) {
		_, err := server.DeleteGraph(ctx, &gripql.GraphID{Graph: schemaName})
		if err != nil {
			return fmt.Errorf("failed to remove previous schema: %v", err)
		}
	}
	_, err := server.AddGraph(ctx, &gripql.GraphID{Graph: schemaName})
	if err != nil {
		return fmt.Errorf("error creating graph '%s': %v", schemaName, err)
	}
	for _, v := range schema.Vertices {
		_, err := server.addVertex(ctx, &gripql.GraphElement{Graph: schemaName, Vertex: v})
		if err != nil {
			return fmt.Errorf("error adding vertex to graph '%s': %v", schemaName, err)
		}
	}
	for _, e := range schema.Edges {
		_, err := server.addEdge(ctx, &gripql.GraphElement{Graph: schemaName, Edge: e})
		if err != nil {
			return fmt.Errorf("error adding edge to graph '%s': %v", schemaName, err)
		}
	}
	return nil
}
