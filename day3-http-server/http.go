package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

/**
 * GeeCache HTTP 服务端
 **/
const defaultBasePath = "/_geecache/"

// HTTPPool 为 HTTP 对等池实现 PeerPicker。
type HTTPPool struct {
	//
	self     string
	basePath string
}

// 初始化一个 HTTP 对等池
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// handle all http request
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
