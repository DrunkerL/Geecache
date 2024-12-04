package geecache

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash     Hash           //哈希函数
	replicas int            //虚拟节点的倍数
	Keys     []int          //哈希环 （排序后的）
	HashMap  map[int]string //虚拟节点与真实节点的映射
}

// New creates a map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		HashMap:  make(map[int]string),
	}
	// 默认的哈希函数
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.Keys = append(m.Keys, hash)
			m.HashMap[hash] = key
		}
	}
	sort.Ints(m.Keys)
}

// Get gets the closest item in the hash to the provided key
func (m *Map) Get(key string) string {
	if len(m.Keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica
	idx := sort.Search(len(m.Keys), func(i int) bool {
		return m.Keys[i] >= hash
	})
	return m.HashMap[m.Keys[idx%len(m.Keys)]]
}
