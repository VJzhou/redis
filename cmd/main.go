package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"redis/model"
	"redis/pkg/gredis"
	"redis/routers"
)

func init () {
	model.Setup()
	gredis.SetUp()
}

func main () {

	gin.SetMode(gin.DebugMode)
	defer model.CloseDb()
	//r.GET("/ping", func(context *gin.Context) {
	//	context.JSON(http.StatusOK, gin.H{
	//		"message" : "hello",
	//	})
	//})
	//
	//r.GET("/shop/:id", func(c *gin.Context) {
	//	id := c.Param("id")
	//
	//	key := fmt.Sprintf("shop:%s", id)
	//	val := redisClient.Get(ctx, key).Val()
	//
	//	if val != "" {
	//		log.Println("query shop cache")
	//		c.JSON(http.StatusOK, gin.H{
	//			"data" : val,
	//		})
	//		return
	//	}
	//
	//	// get lock
	//	lockKey := fmt.Sprintf(lockKeyPrefix + "%s", id)
	//	if !TryLock(lockKey) {
	//		panic("get lock failed")
	//	}
	//	defer LockFree(lockKey)
	//	time.Sleep(time.Millisecond * 50)
	//	// query mysql
	//	row := db.QueryRow("select * from shop where id =?", id)
	//	var shop Shop
	//	if row.Err() != nil {
	//		c.JSON(http.StatusInternalServerError, gin.H{
	//			"message" : "商铺不存在",
	//		})
	//		return
	//	}
	//	log.Println("query mysql")
	//	row.Scan(&shop.Id, &shop.ShopName, &shop.ShopDesc, &shop.ShopAddr, &shop.ShopPhone, &shop.CreatedAt)
	//
	//	// set shop cache
	//	err := redisClient.Set(ctx, key, shop, 0).Err()
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	c.JSON(http.StatusOK, gin.H{
	//		"data1" : shop,
	//	})
	//	return
	//
	//})
	//r.GET("/shop1/:id", func(c *gin.Context) {
	//	id := c.Param("id")
	//	key := fmt.Sprintf("shop:%s", id)
	//	val := redisClient.Get(ctx, key).Val()
	//	//  未查询到缓存
	//	if val == "" {
	//		c.JSON(http.StatusOK, gin.H{
	//			"data" : "",
	//		})
	//		return
	//	}
	//
	//	var shop RedisShop
	//	json.Unmarshal([]byte(val), &shop)
	//	// 判断是否过期
	//	if time.Now().Unix() <  shop.Expire { // 未过期
	//		log.Println("not expire...")
	//		c.JSON(http.StatusOK, gin.H{
	//			"data" : shop,
	//		})
	//		return
	//	}
	//	// 过期
	//	// 获取锁
	//	lockKey := fmt.Sprintf(lockKeyPrefix + "%s", id)
	//	if TryLock(lockKey) {
	//		// 开启独立携程写入redis
	//		go Shop2Redis(db, id)
	//		defer LockFree(lockKey)
	//	}
	//	c.JSON(http.StatusOK, gin.H{
	//		"data" : shop,
	//	})
	//})
	routerInit := routers.InitRouter()

	server := &http.Server{
		Addr: "localhost:9999",
		Handler:routerInit,

	}
	server.ListenAndServe()

}

