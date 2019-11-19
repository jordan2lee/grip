// Code generated by hand-editing-ignore-me-linter. DO NOT EDIT.
// source: gripql.proto

package gripql

//These are custom graph custom statements, which represent operations
//in the traversal that the optimizer may add in, but can't be coded by a
//serialized user request

type GraphStatementLookupVertsIndex struct {
	Labels []string `protobuf:"bytes,1,rep,name=labels" json:"labels,omitempty"`
}

func (*GraphStatementLookupVertsIndex) isGraphStatement_Statement() {}

type GraphStatementEngineCustom struct {
	Desc   string      `protobuf:"bytes,1,opt,name=desc" json:"desc,omitempty"`
	Custom interface{} `protobuf:"bytes,2,opt,name=custom" json:"custom,omitempty"`
}

func (*GraphStatementEngineCustom) isGraphStatement_Statement() {}
