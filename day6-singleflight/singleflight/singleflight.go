package singleflight

import "sync"

/**
 *是 sync 包中的一个类型，主要用于同步多个 goroutine 的执行，
 *确保在主线程或者其他线程继续执行之前，
 *所有相关的 goroutine 都完成了。
 */
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

/**
 * author: Drunker.L:
 * description: Do的作用是，针对相同的key，无论Do被调用多少次，函数fn只会被调用一次，等fn调用结束，返回值或错误
 * param1: key,函数fn
 * returns: 函数fn的执行结果或者错误
 */
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() //如果有请求正在进行中，则阻塞等待
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁
	g.m[key] = c // 添加到g.m中，表明 key 已经有请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() //调用fn发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err // 返回结果
}
