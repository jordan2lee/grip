package engine

import (
	"github.com/bmeg/arachne/badgerdb"
	"github.com/bmeg/arachne/gdbi"
	"github.com/bmeg/arachne/kvi"
	"io/ioutil"
	"os"
)

func NewManager(workDir string) gdbi.Manager {
	return &badgerManager{[]kvi.KVInterface{}, []string{}, workDir}
}

type badgerManager struct {
	kvs     []kvi.KVInterface
	paths   []string
	workDir string
}

func (bm *badgerManager) GetTempKV() kvi.KVInterface {
	td, _ := ioutil.TempDir(bm.workDir, "kvTmp")
	kv, _ := badgerdb.BadgerBuilder(td)

	bm.kvs = append(bm.kvs, kv)
	bm.paths = append(bm.paths, td)
	return kv
}

func (bm *badgerManager) Cleanup() {
	for _, c := range bm.kvs {
		c.Close()
	}
	for _, p := range bm.paths {
		os.RemoveAll(p)
	}
}
