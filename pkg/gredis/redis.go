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

func SetUp () {
	redisClient = NewRedis()
}

func CloseRedis () {
	redisClient.Close()
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