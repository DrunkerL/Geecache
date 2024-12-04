package geecache

import (
	"day5-distributed-node/consistenthash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/**
 * GeeCache HTTP 服务端
 **/
const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool 为 HTTP 对等池实现 PeerPicker。
type HTTPPool struct {
	//
	self        string                 // 记录自己的地址
	basePath    string                 // 节点间通讯的前缀
	mu          sync.Mutex             // 保护并发安全
	peers       *consistenthash.Map    // 一致性哈希算法的Map，用来根据Key来选择节点（远程节点）
	httpGetters map[string]*httpGetter // 映射远程节点与对应的 httpGetter
}

type httpGetter struct {
	baseRUL string
}

// 初始化一个 HTTP 对等池
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

/**
 * 添加和组织远程节点
 * peers: 远程节点的名称
 **/
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)       // 使用默认的参数
	p.peers.Add(peers...)                                    // 添加远程节点
	p.httpGetters = make(map[string]*httpGetter, len(peers)) // 初始化 httpGetters 映射
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseRUL: peer + p.basePath} // 为每个节点创建httpGetter
	}
}

// 打印server的信息
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 远程节点的选择
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 如果找的节点存在并且不是当前的节点，则返回该节点的实例
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil) //类型断言

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf( // 构建请求url
		"%v%v%v",
		h.baseRUL,
		url.QueryEscape(group), // 将传入的字符串进行编码，返回一个安全的 URL 组件
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil) //类型断言

// 处理来自客户端的http请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 判断路径的前缀是否是basePath
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	//<basepath>/<groupname>/<key> required
	/**
	 * 第一个参数：要分割的字符串
	 * 第二个参数：分隔符（这里是 "/"）
	 * 第三个参数：最多分割成几部分（这里是 2）
	 * 返回一个字符串切片
	 **/
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/ovtet-stream")
	w.Write(view.ByteSlice())
}
