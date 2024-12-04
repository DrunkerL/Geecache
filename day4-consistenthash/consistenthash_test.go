package geecache

import (
	"fmt"
	"testing"
)

func TestHashing(t *testing.T) {
	m := New(3, nil)
	m.Add("node1", "node2", "node3")

	fmt.Println("All keys:", m.Keys)
	for _, key := range m.Keys {
		fmt.Printf("Hash: %d, Node: %s\n", key, m.HashMap[key])
	}

	// 测试Get方法
	testKey := "myKey"
	node := m.Get(testKey)
	fmt.Printf("Key '%s' is mapped to node '%s'\n", testKey, node)
}
