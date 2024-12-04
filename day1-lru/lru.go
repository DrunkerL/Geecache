package lru

import "container/list"

//lru 缓存，并发访问不安全

type Cache struct {
	maxBytes  int64                         // 缓存的最大允许使用字节数
	nbytes    int64                         // 当前缓存占用的字节数
	ll        *list.List                    // 维护缓存中元素的顺序
	cache     map[string]*list.Element      // 哈希表旨在O(1)时间内找到key对应的value
	OnEvicted func(key string, value Value) //可选，当一个键值对被移除时调用
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// New is the Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 检查缓存，如果key已经存在，就将其移至队首并更新其value
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry) // 类型断言，将ele转化为entry
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get look up a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)    //如果当前元素被访问，认为最近被使用的概率更高，将元素移到队首
		kv := ele.Value.(*entry) // 将元素转化为entry
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 位于队尾的元素被认为是最不可能被访问的元素，将其删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}

	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
