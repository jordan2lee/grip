package grids

import (
	//"fmt"
	"github.com/akrylysov/pogreb"

	"github.com/bmeg/grip/gdbi"
	"github.com/bmeg/grip/kvi"
	"github.com/bmeg/grip/kvindex"
	"github.com/bmeg/grip/timestamp"
)

// GridsGDB implements the GripInterface using a generic key/value storage driver
type GridsGDB struct {
	keyMap   KeyMap
	keykv    pogreb.DB
	graphkv  kvi.KVInterface
	idx      *kvindex.KVIndex
	ts       *timestamp.Timestamp
}

// GridsGraph implements the GDB interface using a genertic key/value storage driver
type GridsGraph struct {
	kdb      *GridsGDB
	graphID  string
	graphKey uint64
}

// NewKVGraphDB intitalize a new key value graph driver given the name of the
// driver and path/url to create the database at
func NewGridsKVGraphDB(name string, dbPath string) (gdbi.GraphDB, error) {
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		os.Mkdir(dbPath, 0700)
	}
	keyMapPath := fmt.Sprintf("%s/keymap", dbPath)
	graphkvPath := fmt.Sprintf("%s/graph", dbPath)
	idxkvPath := fmt.Sprintf("%s/index", dbPath)
	keyMap, err := pogreb.Open(keyMapPath, nil)
	if err != nil {
		return nil, err
	}
	graphkv, err := badger.NewKVInterface(graphkvPath, nil)
	if err != nil {
		return nil, err
	}
	indexkv, err := badger.NewKVInterface(indexkvPath, nil)
	if err != nil {
		return nil, err
	}
	ts := timestamp.NewTimestamp()
	o := &GridsGDB{keyMap:keyMap, graphkv: graphkv, ts: &ts, idx: kvindex.NewIndex(indexkv)}
	for _, i := range o.ListGraphs() {
		o.ts.Touch(i)
	}
	return o, nil
}

// Close the connection
func (gridb *GridsGDB) Close() error {
	gridb.keyMap.Close()
	gridb.graphkv.Close()
	gridb.indexkv.Close()
	return nil
}
