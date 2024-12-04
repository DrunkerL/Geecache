package geecache

/**
 * 支持并发的缓存
 **/
import (
	"day6-singleflight/singleflight"
	"errors"
	"log"
	"sync"
)

// 							   是
// 接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
//                 |  否                         是
//                 |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
//                             |  否
//                             |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

// 一个Group是一个缓存命名空间及其相关数据，分布加载在多个位置。
type Group struct {
	name      string              // 缓存组的名称
	getter    Getter              // 缓存未命中时的回调函数
	mainCache cache               // 实际缓存存储数据的地方
	peers     PeerPicker          // 远程缓存节点
	loader    *singleflight.Group // 确保相同的key只被请求一次
}

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

/**
 * author: Drunker.L:
 * description: 获取指定key的value值。先在本地缓存中找，缓存命中返回指定value
 * 				缓存未命中，调用load方法，从远程节点中获取
 */
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.load(key)
}

// 使用 PickPeer() 方法选择节点，若非本机节点，则调用 getFromPeer() 从远程获取。
// 若是本机节点或失败，则回退到 getLocally()
func (g *Group) load(key string) (value ByteView, err error) {
	// 每一个key都被请求一次，无论是本地还是远程，不管请求了多少次
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return

}

/*
*
  - author: Drunker.L:
  - description: 远程节点未命中的情况下，调用此方法
    g.getter实现缓存未命中时的回调函数，一般是加载数据库
*/
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

/**
 * author: Drunker.L:
 * description: 使用实现了 PeerGetter 接口的 httpGetter 从访问远程节点，获取缓存值
 * param1: peer 代表远程节点
 */
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

/**
 * author: Drunker.L:
 * description: 将次 key-value对，放进当前缓存中，以便下次访问（lru思想）
 */
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 将实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
