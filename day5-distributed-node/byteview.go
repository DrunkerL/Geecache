package geecache

// ByteView持有一个字节的不可变视图
/**
 * ByteView 只有一个数据成员，b []byte，
 * b 将会存储真实的缓存值。
 * 选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。
 */
type ByteView struct {
	b []byte
}

// len return the length of view
/**
 * 实现 Len() int 方法，我们在 lru.Cache 的实现中，
 * 要求被缓存对象必须实现 Value 接口，即 Len() int 方法，返回其所占的内存大小。
 */
func (v ByteView) Len() int {
	return len(v.b)
}

/**
 * b 是只读的，使用 ByteSlice() 方法返回一个拷贝，防止缓存值被外部程序修改。
 */
func (v *ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

/**
 * author: Drunker.L:
 * description: 返回数据的一份拷贝，防止本地数据（真实数据）被修改
 */
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// String returns the data as a string, making a copy if necessary.
func (v *ByteView) String() string {
	return string(v.b)
}
