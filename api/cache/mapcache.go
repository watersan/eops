package cache

import (
	//"encoding/gob"
	"fmt"
	//"io"
	//"os"
	//"runtime"
	"sync"
	"time"
)

//Item 缓存的存储结构
type Item struct {
	Object     interface{}
	Expiration int64
}

//MapCache 内存缓存结构体
type MapCache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	//	janitor           *janitor
}

const (
	//NoExpiration For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	//DefaultExpiration For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 5
)

//NewMapCache 创建缓存对象
func NewMapCache() *MapCache {
	items := make(map[string]Item)
	mc := &MapCache{defaultExpiration: DefaultExpiration, items: items}
	return mc
}

//Expired Returns true if the item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

//Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *MapCache) Add(args ...interface{}) error {
	k := args[0].(string)
	x := args[1]
	var d time.Duration
	if len(args) >= 3 {
		d = args[2].(time.Duration)
	}
	_, found := c.Get(k)
	if found != nil {
		return fmt.Errorf("Item %s already exists", k)
	}
	c.Set(k, x, d)
	return nil
}

// Set an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *MapCache) Set(args ...interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var e int64
	var d time.Duration
	d = 0
	k := args[0].(string)
	x := args[1]
	if len(args) >= 3 {
		d = args[2].(time.Duration)
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
	return nil
}

//SetExpire 设置过期时间
func (c *MapCache) SetExpire(key string, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var e int64
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	v, found := c.items[key]
	if !found || v.Expired() {
		return fmt.Errorf("Item %s not found", key)
	}
	v.Expiration = e
	c.items[key] = v
	return nil
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *MapCache) Get(k string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		return nil, fmt.Errorf("Item %s not found", k)
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, fmt.Errorf("Item %s is expired", k)
		}
	}
	return item.Object, nil
}

// Increment an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *MapCache) Increment(k string, n uint64) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, found := c.items[k]
	if !found || v.Expired() {
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint64)
	if !ok {
		return 0, fmt.Errorf("The value for %s is not an int", k)
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	return nv, nil
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (c *MapCache) Decrement(k string, n uint64) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, found := c.items[k]
	if !found || v.Expired() {
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint64)
	if !ok {
		return 0, fmt.Errorf("The value for %s is not an int", k)
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	return nv, nil
}

//Del 设置过期时间
func (c *MapCache) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, found := c.items[key]; found {
		delete(c.items, key)
	}
}

//GetAll 获取所有缓存内容
func (c *MapCache) GetAll() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items
}

//Clean 清理过期内容
func (c *MapCache) Clean() (int, int) {
	var allCount, delCount int
	for k, item := range c.items {
		allCount++
		if item.Expired() {
			c.Del(k)
			delCount++
		}
	}
	return allCount, delCount
}
