package connection

import "sync"

// KeyValueContext：键值对上下文
type KeyValueContext struct {
	// 读写锁
	mu sync.RWMutex
	// map：键为 string，值为接口 interface{}
	kv map[string]interface{}
}

// Set：进行 KeyValueContext 键值设置
func (c *KeyValueContext) Set(key string, value interface{}) {
	// 可并发读，不可并发写
	c.mu.Lock()
	if c.kv == nil {
		c.kv = make(map[string]interface{})
	}
	// 进行元素键值赋值
	c.kv[key] = value
	// 解锁
	c.mu.Unlock()
}

// Delete： 删除 map 中的映射
func (c *KeyValueContext) Delete(key string) {
	c.mu.Lock()
	delete(c.kv, key)
	c.mu.Unlock()
}

// Get： 根据键 key 得到对应的 value 及是否存在标志 bool
func (c *KeyValueContext) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.kv[key]
	c.mu.RUnlock()
	return
}

// reset： 进行 map 的重置，所有的内容清空
func (c *KeyValueContext) reset() {
	c.mu.Lock()
	c.kv = nil
	c.mu.Unlock()
}
