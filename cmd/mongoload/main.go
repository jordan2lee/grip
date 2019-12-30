package mongoload

import (
	"fmt"
	"context"
	"time"
	"sync"
	//"io"
	//"strings"

	"github.com/bmeg/grip/log"
	"github.com/bmeg/grip/mongo"
	"github.com/bmeg/grip/util"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/spf13/cobra"
)

var mongoHost = "localhost"
var database = "gripdb"

var graph string
var vertexFile string
var edgeFile string

var batchSize = 1000
var maxRetries = 3

func found(set []string, val string) bool {
	for _, i := range set {
		if i == val {
			return true
		}
	}
	return false
}

// MaxRetries is the number of times driver will reconnect on connection failure
// TODO, move to per instance config, rather then global
var MaxRetries = 3

/*
func isNetError(e error) bool {
	if e == io.EOF {
		return true
	}
	if b, ok := e.(*mgo.BulkError); ok {
		for _, c := range b.Cases() {
			if c.Err == io.EOF {
				return true
			}
			if strings.Contains(c.Err.Error(), "connection") {
				return true
			}
		}
	}
	return false
}
*/

func boolPtr(a bool) *bool {
	return &a
}

func vertexCollection(session *mgo.Client, database string, graph string) *mgo.Collection {
	return session.Database(database).Collection(fmt.Sprintf("%s_vertices", graph))
}

func edgeCollection(session *mgo.Client, database string, graph string) *mgo.Collection {
	return session.Database(database).Collection(fmt.Sprintf("%s_edges", graph))
}

func addGraph(client *mgo.Client, database string, graph string) error {
	graphs := client.Database(database).Collection("graphs")
	_, err := graphs.InsertOne(context.Background(), bson.M{"_id": graph})
	if err != nil {
		return fmt.Errorf("failed to insert graph %s: %v", graph, err)
	}

	e := edgeCollection(client, database, graph)
	eiv := e.Indexes()
	_, err = eiv.CreateOne(
		context.Background(),
		mgo.IndexModel{
			Keys: []string{"from"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	_, err = eiv.CreateOne(
		context.Background(),
		mgo.IndexModel{
			Keys: []string{"to"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	_, err = eiv.CreateOne(
		context.Background(),
		mgo.IndexModel{
			Keys: []string{"label"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	v := vertexCollection(client, database, graph)
	viv := v.Indexes()
	_, err = viv.CreateOne(
		context.Background(),
		mgo.IndexModel{
			Keys: []string{"label"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}
	return nil
}

func docWriter(col *mgo.Collection, docChan chan bson.M, sn *sync.WaitGroup) {
	defer sn.Done()
	docBatch := make([]mgo.WriteModel, 0, batchSize)
	for ent := range docChan {
		i := mgo.NewInsertOneModel()
		i.SetDocument(ent)
		docBatch = append(docBatch, i)
		if len(docBatch) > batchSize {
			_, err := col.BulkWrite(context.Background(), docBatch)
			if err != nil {
				log.Errorf("%s", err)
			}
			docBatch = make([]mgo.WriteModel, 0, batchSize)
		}
	}
	if len(docBatch) > 0 {
		col.BulkWrite(context.Background(), docBatch)
	}
}


// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "mongoload <graph>",
	Short: "Directly load data into mongodb",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if vertexFile == "" && edgeFile == "" {
			return fmt.Errorf("no edge or vertex files were provided")
		}

		graph = args[0]


		// Connect to mongo and start the bulk load process
		log.Infof("Loading data into graph: %s", graph)
		client, err := mgo.NewClient(options.Client().ApplyURI(mongoHost))
		if err != nil {
			return err
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)

		addGraph(client, database, graph)

		vertexCol := client.Database(database).Collection(fmt.Sprintf("%s_vertices", graph))
		edgeCol := client.Database(database).Collection(fmt.Sprintf("%s_edges", graph))

		vertexDocChan := make(chan bson.M, 5)
		edgeDocChan := make(chan bson.M, 5)

		s := &sync.WaitGroup{}
		go docWriter(edgeCol, edgeDocChan, s)
		s.Add(1)
		go docWriter(vertexCol, vertexDocChan, s)
		s.Add(1)

		if vertexFile != "" {
			log.Infof("Loading vertex file: %s", vertexFile)

			bulkVertChan := make(chan []map[string]interface{}, 5)
			docBatch := make([]map[string]interface{}, 0, batchSize)

			go func() {
				count := 0
				for batch := range bulkVertChan {
					for _, data := range batch {
						vertexDocChan <- data
						if count%1000 == 0 {
							log.Infof("Loaded %d vertices", count)
						}
					}
				}
				close(vertexDocChan)
				log.Infof("Loaded %d vertices", count)
			}()

			vertChan, err := util.StreamVerticesFromFile(vertexFile)
			if err != nil {
				return err
			}
			for v := range vertChan {
				data := mongo.PackVertex(v)
				docBatch = append(docBatch, data)
				if len(docBatch) > batchSize {
					bulkVertChan <- docBatch
					docBatch = make([]map[string]interface{}, 0, batchSize)
				}
			}
			if len(docBatch) > 0 {
				bulkVertChan <- docBatch
			}
			close(bulkVertChan)
		}

		if edgeFile != "" {
			log.Infof("Loading edge file: %s", edgeFile)

			bulkEdgeChan := make(chan []map[string]interface{}, 5)
			docBatch := make([]map[string]interface{}, 0, batchSize)

			go func() {
				count := 0
				for batch := range bulkEdgeChan {
					for _, data := range batch {
						edgeDocChan <- data
						if count%1000 == 0 {
							log.Infof("Loaded %d edges", count)
						}
					}
				}
				log.Infof("Loaded %d edges", count)
			}()

			edgeChan, err := util.StreamEdgesFromFile(edgeFile)
			if err != nil {
				return err
			}
			for e := range edgeChan {
				data := mongo.PackEdge(e)
				//if data["_id"] == "" {
				//	data["_id"] = bson.NewObjectId().Hex()
				//}
				docBatch = append(docBatch, data)
				if len(docBatch) > batchSize {
					bulkEdgeChan <- docBatch
					docBatch = make([]map[string]interface{}, 0, batchSize)
				}
			}
			if len(docBatch) > 0 {
				bulkEdgeChan <- docBatch
			}
			close(bulkEdgeChan)
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&mongoHost, "mongo-host", mongoHost, "mongo server url")
	flags.StringVar(&database, "database", database, "database name in mongo to store graph")
	flags.StringVar(&vertexFile, "vertex", "", "vertex file")
	flags.StringVar(&edgeFile, "edge", "", "edge file")
	flags.IntVar(&batchSize, "batch-size", batchSize, "mongo bulk load batch size")
}
