package cache

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	//"fmt"
	//"reflect"
)

//RedisCache redis缓存结构
type RedisCache struct {
	pool *redis.Pool
}

//redisInit Redis连接池
func RedisInit(server, password string) *RedisCache {
	rcache := &RedisCache{}
	rcache.pool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   100,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if _, autherr := c.Do("AUTH", password); autherr != nil {
				c.Close()
				return nil, autherr
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return rcache
}

//Get 获取key的值
func (rcache *RedisCache) Get(key string) (interface{}, error) {
	conn := rcache.pool.Get()
	reply, err := conn.Do("GET", key)
	//value, err = redis.String(reply, err)
	return reply, err
}

//Set 为key设置新值
func (rcache *RedisCache) Set(args ...interface{}) error {
	conn := rcache.pool.Get()
	var reply interface{}
	var err error
	reply, err = conn.Do("SET", args...)
	if reply == nil {
		err = errors.New("set failed")
	}
	return err
}

//Add 为key设置新值
func (rcache *RedisCache) Add(args ...interface{}) error {
	conn := rcache.pool.Get()
	var reply interface{}
	var err error
	reply, err = conn.Do("SETNX", args...)
	if reply == nil {
		err = errors.New("add failed")
	}
	return err
}
