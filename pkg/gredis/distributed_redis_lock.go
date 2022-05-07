package gredis

import (
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
	Durations time.Duration
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

func (drl *disRedisLock) LockFree () {
	err := redisClient.Del(ctx, drl.Key).Err()
	if err != nil {
		log.Println(err.Error())
	}
}






