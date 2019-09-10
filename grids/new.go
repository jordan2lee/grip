package grids

import (
	"fmt"
	"os"
	"github.com/akrylysov/pogreb"

	"github.com/bmeg/grip/gdbi"
	"github.com/bmeg/grip/kvi"
	"github.com/bmeg/grip/kvi/badgerdb"
	"github.com/bmeg/grip/kvindex"
	"github.com/bmeg/grip/timestamp"

	log "github.com/sirupsen/logrus"
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

// NewKVGraphDB intitalize a new grids graph driver
func NewGridsGraphDB(dbPath string) (gdbi.GraphDB, error) {
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		os.Mkdir(dbPath, 0700)
	}
	keykvPath := fmt.Sprintf("%s/keymap", dbPath)
	graphkvPath := fmt.Sprintf("%s/graph", dbPath)
	//indexkvPath := fmt.Sprintf("%s/index", dbPath)
	keykv, err := pogreb.Open(keykvPath, nil)
	if err != nil {
		return nil, err
	}
	graphkv, err := badgerdb.NewKVInterface(graphkvPath, kvi.Options{})
	if err != nil {
		return nil, err
	}
	/*
	indexkv, err := badgerdb.NewKVInterface(indexkvPath, kvi.Options{})
	if err != nil {
		return nil, err
	}
	*/
	ts := timestamp.NewTimestamp()
	o := &GridsGDB{keyMap:NewKeyMap(keykv), graphkv: graphkv, ts: &ts, idx: kvindex.NewIndex(graphkv)}
	for _, i := range o.ListGraphs() {
		o.ts.Touch(i)
	}
	log.Printf("Starting GRIDS driver")
	return o, nil
}

// Close the connection
func (gridb *GridsGDB) Close() error {
	gridb.keyMap.Close()
	gridb.graphkv.Close()
	//gridb.indexkv.Close()
	return nil
}
