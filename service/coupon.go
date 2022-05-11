package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"redis/model"
	"redis/pkg/gredis"
	"strconv"
	"sync"
	"time"
)

func init () {
	//go func() {
	//	select {
	//	case c := <- orderInfoChan :
	//		go c.AddSeckillCouponOrderByDrl()
	//	}
	//}()

	cc.Key = "coupon-order"
	cc.GroupName = "order-group"
	// 从最后一条开始
	cc.Start = "$"
	cc.ConsumerName = "consumer1"
	cc.Streams = []string{cc.Key, ">"}
	cc.ConsumerCount = 1
	// 创建消费组
	cc.CreateStreamsGroup()
}

var (
	wg sync.WaitGroup
	cc CouponStreams
)


type couponOrderToQ struct {
	OrderId string // 订单ID
	Uid uint64 // 用户ID
	Cid uint64 // 优惠券ID
}
//var orderInfoChan = make(chan *couponOrderToQ)

type CouponStreams struct {
	Key string
	GroupName string
	ConsumerName string
	Streams []string
	ConsumerCount int
	IsCreate bool
	Start string
}

func (cc *CouponStreams) CreateStreamsGroup () {
	gredis.CreateStreamGroup(cc.Key, cc.GroupName, cc.Start, true)
}

func (cc *CouponStreams) ProcessStream (stream redis.XMessage) {
	defer wg.Done()
	var c couponOrderToQ
	c.OrderId = stream.Values["order-id"].(string)
	c.Uid, _= strconv.ParseUint(stream.Values["uid"].(string), 10, 64)
	c.Cid, _= strconv.ParseUint(stream.Values["cid"].(string), 10, 64)
	c.Cid, _= strconv.ParseUint(stream.Values["cid"].(string), 10, 64)
	ret, err := gredis.MessageAck(cc.Key, cc.GroupName, stream.ID)
	log.Println("ack ret: ", ret)
	if err == nil {
		log.Println("insert coupon order ")
		c.AddSeckillCouponOrderByDrl()
	} else {
		log.Println("ack failed")
	}
}


func (cc *CouponStreams) ConsumerEvents () {
	for  {
		// 消费
		sliceCmd :=gredis.ConsumerGroup(cc.GroupName, cc.ConsumerName, cc.Streams, int64(cc.ConsumerCount))
		if sliceCmd.Err() != nil {
			log.Println("xread group failed: ",sliceCmd.Err())
		}
		sliceEntries := sliceCmd.Val()
		for _, stream := range sliceEntries[0].Messages {
			wg.Add(1)
			go cc.ProcessStream(stream)
		}
		wg.Wait()
	}
}

func (cc *CouponStreams) ConsumerPendingEvents () {
	for {
		// 将消息从pending 列表取出
		var streamRetry []string
		pendingStreams, err :=gredis.ConsumerPendingGroup(cc.Key, cc.GroupName, "0", "+", int64(cc.ConsumerCount))
		if err != nil {
			log.Println(err.Error())
		}

		for _, stream := range pendingStreams {
			streamRetry = append(streamRetry, stream.ID)
		}
		// 重新认领消息
		if len(streamRetry) > 0 {
			sliceCmd := gredis.ClaimMessage(cc.Key, cc.GroupName, cc.ConsumerName, time.Second * 30,streamRetry )
			streams, err := sliceCmd.Result()
			if err != nil {
				log.Println("claim message error: ", err.Error())
			}
			// 重新消费消息
			for _, stream := range streams {
				wg.Add(1)
				go cc.ProcessStream(stream)
			}
		}

	}
}

func NewCC  () *CouponStreams {
	return &cc
}


func NewCouponOrderToQ (orderid string, cid, uid uint64) *couponOrderToQ{
	return &couponOrderToQ{
		OrderId: orderid,
		Uid:     uid,
		Cid:     cid,
	}
}

func AddCouponOrder (cid , uid int) (int64, error) {
	db := model.NewDB()

	tx, err := db.Begin()
	var co model.CouponOrder
	co.Cid = uint64(cid)
	co.Uid = uint64(uid)

	result2, err := model.DecrByStep(cid, 1)
	affectRow, _ := result2.RowsAffected()
	if affectRow == 0 || err != nil{
		_ = tx.Rollback()
		return 0, errors.New("库存不足")
	}

	result1, err :=model.AddCouponOrder(&co)
	insertId , _ := result1.LastInsertId()
	if insertId < 0 || err != nil{
		_ = tx.Rollback()
		return 0, errors.New("添加订单失败")
	}
	_ =tx.Commit()
	return insertId , nil
}

func IsExistsCo (cid, uid int ) bool {
	var co model.CouponOrder
	row := model.GetCOByUidAndCid(cid, uid)
	row.Scan(&co.Id, &co.Status, &co.CreatedAt, &co.Cid, &co.Uid)
	return co.Id > 0
}

// 分布式锁
func DoAddSeckillCouponOrder (ctx *gin.Context) (int64, error){
	sid := ctx.PostForm("id")
	id, _:=strconv.Atoi(sid)
	sc , err := model.GetSCRow(id)
	if err != nil {
		log.Println(err.Error())
		return 0,err
	}
	if sc == nil {
		return 0, errors.New("秒杀优惠券不存在")
	}
	begin, err:= model.ParseStringToTime(sc.BeginTime)
	if err!= nil {
		log.Println("string to time failed")
		return 0,errors.New("添加失败")
	}

	if time.Now().Before(begin) {
		return 0,errors.New("活动未开始")
	}

	end, err:= model.ParseStringToTime(sc.EndTime)
	if err!= nil {
		log.Println("string to time failed")
		return 0,errors.New("请重试")
	}
	if time.Now().After(end) {
		return 0, errors.New("活动已结束")
	}

	lockKey := fmt.Sprintf(gredis.COUPON_UID_LOCK, 2)
	uid := rand.Intn(200)
	drl := gredis.NewDisRedisLock(lockKey, uuid.NewString(), time.Second * 2)
	if !drl.TryLock() {
		return 0, errors.New("请重试")
	}
	isExists := IsExistsCo(2, 2)
	// 检查是否已下单
	if isExists {
		return 0, errors.New("请勿重复下单")
	}
	insertId, err := AddCouponOrder(id,uid)
	drl.AtomLockFree()
	if err != nil {
		return 0, err
	}
	return insertId, nil
}

// 接口时间优化
// 1 将券库存数量添加redis
// 2 将已购买的人放入redis 判断
// 3 将1、2 使用redis脚本判断
// 4 将订单放入队列， 使用其他线程处理
func DoAddSeckillCouponOrderV1 (ctx *gin.Context) (int64, error){
	sid := ctx.PostForm("id")
	id, _:=strconv.Atoi(sid)
	sc , err := model.GetSCRow(id)
	if err != nil {
		log.Println(err.Error())
		return 0,err
	}
	if sc == nil {
		return 0, errors.New("秒杀优惠券不存在")
	}
	begin, err:= model.ParseStringToTime(sc.BeginTime)
	if err!= nil {
		log.Println("string to time failed")
		return 0,errors.New("添加失败")
	}

	if time.Now().Before(begin) {
		return 0,errors.New("活动未开始")
	}

	end, err:= model.ParseStringToTime(sc.EndTime)
	if err!= nil {
		log.Println("string to time failed")
		return 0,errors.New("请重试")
	}
	if time.Now().After(end) {
		return 0, errors.New("活动已结束")
	}

	var luaScript = redis.NewScript(`
		-- get stock
		local stock = redis.call('get', KEYS[1])
		if (tonumber(stock)) <= 0 then
			-- 库存不足
			return 1
		end
		
		if redis.call('sismember', KEYS[2], ARGV[1]) == 1 then
			-- 用户已下单
			return 2
		end
		
		-- 库存-1
		redis.call('incrby', KEYS[1], -1)
		-- 添加用户ID
		redis.call('sadd', KEYS[2], ARGV[1])
		-- 添加到消息队列
		redis.call('xadd', 'coupon-order', '*', 'order-id', ARGV[2], 'uid', ARGV[1], 'cid', ARGV[3])
		return 0
	`)

	couponStockKey := fmt.Sprintf("coupon:stock:%d", 2)
	couponSeckillUidKey := "coupon:seckill:2"

	rand.Seed(time.Now().UnixNano())
	orderId := rand.Intn(200)
	uid := strconv.Itoa(rand.Intn(200))

	keys := []string{couponStockKey, couponSeckillUidKey}
	argv := []string{uid, strconv.Itoa(orderId), "2"}
	res, err := gredis.RunLuaScript(luaScript, keys, argv)
	if err != nil {
		log.Println(err.Error())
		return 0, errors.New("请重试")
	}

	switch res.(int64) {
	case 1:
		return 0, errors.New("库存不足")
	case 2:
		return 0, errors.New("请勿重复下单")
	}
	return int64(orderId), nil
}

func (c *couponOrderToQ) AddSeckillCouponOrderByDrl () {
	lockKey := fmt.Sprintf(gredis.COUPON_UID_LOCK, c.Uid)
	drl := gredis.NewDisRedisLock(lockKey, uuid.NewString(), time.Second * 5)
	if !drl.TryLock() {
		log.Println("获取锁失败")
	}
	isExists := IsExistsCo(int(c.Cid), int(c.Uid))
	// 检查是否已下单
	if isExists {
		log.Println("请勿重复下单")
	}
	_, err := AddCouponOrder(int(c.Cid) ,int(c.Uid))
	drl.AtomLockFree()
	if err != nil {
		// todo 添加失败，重试？
		log.Println("添加订单失败")
	}
}