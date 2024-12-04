package geecache

// 通过 PeerPicker 实现节点的选择，决定哪个节点负责处理特定的 key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 通过 PeerGetter 实现节点间的通信，从远程节点获取缓存数据
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
