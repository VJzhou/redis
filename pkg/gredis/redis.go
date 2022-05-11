package gredis

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"redis/model"
	"time"
)

var (
	redisClient *redis.Client
	ctx = context.Background()
)

func NewRedis () *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})
}

func init () {
	SetUp()
}

func SetUp () {
	redisClient = NewRedis()
}

func CloseRedis () {
	redisClient.Close()
}

func RunLuaScript (s *redis.Script, keys []string, argv interface{}) (interface{}, error) {
	return  s.Run(ctx,redisClient, keys, argv).Result()
}

func NewOrderId () {

}

func CreateStreamGroup (key, groupName, start string, mkStream bool) {
	var err error
	if mkStream {
		err = redisClient.XGroupCreateMkStream(ctx, key, groupName, start).Err()
	} else {
		err = redisClient.XGroupCreate(ctx, key, groupName, start).Err()
	}
	if err != nil {
		log.Println("create group failed: ",err.Error())
	}
}

func ConsumerGroup (groupName , consumerName string, streams []string, consumerCount int64) *redis.XStreamSliceCmd{
	return redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  streams,
		Count:    consumerCount,
		Block:    0,
		NoAck:    false,
	})

}

func ConsumerPendingGroup (key, groupName, start, end string, count int64 ) ([]redis.XPendingExt, error) {
	return redisClient.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   key,
		Group:    groupName,
		Start:    start,
		End:      end,
		Count:    count,
	}).Result()
}

func ClaimMessage (key, groupName, consumer string, mi time.Duration, messages []string ) *redis.XMessageSliceCmd{
	return redisClient.XClaim(ctx, &redis.XClaimArgs{
		Stream:   key,
		Group:    groupName,
		Consumer: consumer,
		MinIdle:  mi,
		Messages: messages,
	})
}

func MessageAck (key, groupName, messageId string ) (int64, error) {
	return  redisClient.XAck(ctx, key,groupName, messageId).Result()
}

func Shop2Redis (db *sql.DB, shopID string) (*model.RedisShop, error) {
	var shop model.RedisShop
	time.Sleep(time.Millisecond * 200)
	row := db.QueryRow("select * from shop where id = ?", shopID)
	if row.Err() != nil {
		log.Println("query failed")
	}
	row.Scan(&shop.Id, &shop.ShopName, &shop.ShopDesc, &shop.ShopAddr, &shop.ShopPhone, &shop.CreatedAt)
	shop.Expire = time.Now().Unix() + 20

	key := fmt.Sprintf("shop:%s", shopID)
	return  &shop, redisClient.Set(ctx, key, shop, 0).Err()
}