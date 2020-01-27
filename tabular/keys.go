package tabular


import (
	"encoding/binary"
)

var pathCount  = []byte("c")
var pathPrefix = []byte("p")
var linePrefix = []byte("l")
var idPrefix   = []byte("i")
var countPrefix = []byte("n")
var idxPrefix   = []byte("x")

func PathNumKey() []byte {
	return pathCount
}

// PathKey produces the byte key for a particular file path
func PathKey(path string) []byte {
  p := []byte(path)
  out := make([]byte, 1 + len(p))
  p[0] = pathPrefix[0]
  for i := 0; i < len(p); i++ {
    out[i+1] = p[i]
  }
  return out
}


// LineKey
func LineKey(pathID, line uint64) []byte {
  out := make([]byte, 1+binary.MaxVarintLen64*2)
	out[0] = linePrefix[0]
	binary.PutUvarint(out[1:binary.MaxVarintLen64+1], pathID)
	binary.PutUvarint(out[binary.MaxVarintLen64+1:2*binary.MaxVarintLen64+1], line)
  return out
}

func IDPrefix(pathID uint64) []byte {
  out := make([]byte, 1 + binary.MaxVarintLen64)
  out[0] = idPrefix[0]
	binary.PutUvarint(out[1:binary.MaxVarintLen64+1], pathID)
  return out
}

// IDKey produces the byte key for a particular file path
func IDKey(pathID uint64, id string) []byte {
  p := []byte(id)
  out := make([]byte, 1 + binary.MaxVarintLen64 + len(p))
  out[0] = idPrefix[0]
	binary.PutUvarint(out[1:binary.MaxVarintLen64+1], pathID)
  for i := 0; i < len(p); i++ {
    out[i+1+binary.MaxVarintLen64] = p[i]
  }
  return out
}

func IDKeyParse(key []byte) (uint64, []byte) {
	pathID, _ := binary.Uvarint(key[1:binary.MaxVarintLen64+1])
	sLen := len(key) - (binary.MaxVarintLen64+1)
	id := make([]byte, sLen)
	for i := 0; i < sLen; i++ {
		id[i] = key[i+1+binary.MaxVarintLen64]
	}
	return pathID, id
}

func LineCountKey(pathID uint64) []byte {
	out := make([]byte, 1+binary.MaxVarintLen64)
	out[0] = countPrefix[0]
	binary.PutUvarint(out[1:binary.MaxVarintLen64+1], pathID)
  return out
}

func IndexKey(pathID uint64, column uint64, value string, lineNum uint64) []byte {
	s := []byte(value)
	out := make([]byte, 1+(binary.MaxVarintLen64*3)+len(s))
	out[0] = idxPrefix[0]
	binary.PutUvarint(out[1:binary.MaxVarintLen64+1], pathID)
	binary.PutUvarint(out[1+binary.MaxVarintLen64:2*binary.MaxVarintLen64+1], column)
	for i := 0; i < len(s); i++ {
		out[i+1+2*binary.MaxVarintLen64] = value[i]
	}
	binary.PutUvarint(out[1+2*binary.MaxVarintLen64+len(s):1+3*binary.MaxVarintLen64+len(s)], lineNum)
	return out
}

func IndexKeyParse(key []byte) (uint64, uint64, string, uint64) {
	pathID, _ := binary.Uvarint(key[1:binary.MaxVarintLen64+1])
	column, _ := binary.Uvarint(key[1+binary.MaxVarintLen64:1+2*binary.MaxVarintLen64])
	sLen := len(key) - (3*binary.MaxVarintLen64+1)
	value := make([]byte, sLen)
	for i := 0; i < sLen; i++ {
		value[i] = key[i+1+2*binary.MaxVarintLen64]
	}
	line, _ := binary.Uvarint(key[1+2*binary.MaxVarintLen64+sLen:1+3*binary.MaxVarintLen64+sLen])
	return pathID, column, string(value), line
}
