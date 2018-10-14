package cache

//Cacher 针对缓存的接口，提供缓存所需的方法
type Cacher interface {
	Get(key string) (interface{}, error)
	Set(args ...interface{}) error
	Add(args ...interface{}) error
}

//New 创建新的缓存对象
func New(app string, server, password string) Cacher {
	//cache := &Cache{}
	switch app {
	case "redis":
		redis := RedisInit(server, password)
		return redis
	case "map":
		mapcache := NewMapCache()
		return mapcache
	}
	return nil
}
