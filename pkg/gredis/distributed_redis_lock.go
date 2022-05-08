package gredis

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

var (
	COUPON_UID_LOCK = "coupon:lock:%d"
)

// 分布式redis锁
type disRedisLock struct {
	Key string
	Val interface{}
	Durations time.Duration // 过期时间
}

func NewDisRedisLock (key string, val interface{}, duration time.Duration) *disRedisLock{
	return &disRedisLock{
		Key:       key,
		Val:       val,
		Durations: duration,
	}
}

// 上锁
func (drl *disRedisLock) TryLock () (b bool) {
	log.Println("lock key: ", drl.Key)
	b, err := redisClient.SetNX(ctx, drl.Key, drl.Val, drl.Durations).Result()
	if err != nil {
		log.Println(err.Error())
	}
	return
}

// 释放锁
// 锁误删除问题
// 1。 删除的不是自己的锁， 解决方案，在key-》val 为随机值，删除锁时判断当前线程的val与get key 的val 是否相等， 相等删除key
//redis锁
//
//线程1--------获取锁ok-----执行业务，产生阻塞，锁超时自动释放--------------------阻塞结束，释放锁（此时释放了线程2锁）
//
//线程2-------------------------------------------------- 获取锁----执行业务---
//																			此时线程就不再是串行执行了
//线程3---------------------------------------------------------------------获取锁开始业务--
func (drl *disRedisLock) LockFree () {
	val, err := redisClient.Get(ctx, drl.Key).Result()
	if err != nil {
		log.Println(err.Error())
		return
	}
	if val == drl.Val { // 只能删除自己的锁
		err := redisClient.Del(ctx, drl.Key).Err()
		if err != nil {
			log.Println(err.Error())
		}
	}
}

//2 判断锁一致和释放锁不是 原子操作，解决方案：使用lua 脚本，原子性操作
//redis锁
//
//线程1---获取锁ok---执行业务----判断锁是自己的，释放锁时阻塞了，锁超时释放-----阻塞恢复了，删除锁（线程2的锁）
//
//线程2------------------------获取锁成功------------------执行业务---------------------------------
//																						此时线程就不再是串行执行了
//线程3-------------------------------------------------------------------------------------------获取锁ok-----执行
func (drl *disRedisLock) AtomLockFree () {
	var luaScript = redis.NewScript(`
		local getVal = redis.call('get', KEY[1])
		local val = ARGV[1]
		if getVal == val then
			redis.call("del", KEYS[1])
		end
    `)

	keys := []string{drl.Key}
	argv := []string{fmt.Sprint(drl.Val)}
	err := luaScript.Run(ctx, redisClient, keys, argv).Err()
	if err != nil {
		log.Println(err.Error())
	}
}

// todo 分段锁->解决分布式锁的性能
// 业务逻辑执行时间20ms -> 一秒只能处理50个请求





