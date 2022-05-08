package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"redis/model"
	"redis/pkg/app"
	"redis/pkg/gredis"
	"redis/service"
	"strconv"
	"time"
)

var (
	isExists bool
	//mutex sync.Mutex
)

// 单机 锁
//func AddCouponOrder (ctx *gin.Context) {
//	appG := app.Gin{C:ctx}
//	sid := ctx.PostForm("id")
//	id, _:=strconv.Atoi(sid)
//	sc , err := model.GetSCRow(id)
//	if err != nil {
//		log.Println(err.Error())
//	}
//	if sc == nil {
//		appG.Response(http.StatusBadRequest, "秒杀优惠券不存在", struct{}{})
//		return
//	}
//	begin, err:= model.ParseStringToTime(sc.BeginTime)
//	if err!= nil {
//		log.Println("string to time failed")
//	}
//
//	if time.Now().Before(begin) {
//		appG.Response(http.StatusBadRequest, "活动未开始",  struct{}{})
//		return
//	}
//
//	end, err:= model.ParseStringToTime(sc.EndTime)
//	if err!= nil {
//		log.Println("string to time failed")
//	}
//	if time.Now().After(end) {
//		appG.Response(http.StatusBadRequest, "活动已结束",  struct{}{})
//		return
//	}
//
//	mutex.Lock()
//	isExists = service.IsExistsCo(2, 2)
//	// 检查是否已下单
//	if isExists {
//		appG.Response(http.StatusBadRequest,"请勿重复下单", struct {}{})
//		mutex.Unlock()
//		return
//	}
//	insertId, err := service.AddCouponOrder(id)
//	mutex.Unlock()
//	if err != nil {
//		appG.Response(http.StatusBadRequest, err.Error(),  struct{}{})
//		return
//	}
//	appG.Response(http.StatusOK, "",  insertId)
//}

// 分布式锁
func AddCouponOrder (ctx *gin.Context) {
	appG := app.Gin{C:ctx}
	sid := ctx.PostForm("id")
	id, _:=strconv.Atoi(sid)
	sc , err := model.GetSCRow(id)
	if err != nil {
		log.Println(err.Error())
	}
	if sc == nil {
		appG.Response(http.StatusBadRequest, "秒杀优惠券不存在", struct{}{})
		return
	}
	begin, err:= model.ParseStringToTime(sc.BeginTime)
	if err!= nil {
		log.Println("string to time failed")
	}

	if time.Now().Before(begin) {
		appG.Response(http.StatusBadRequest, "活动未开始",  struct{}{})
		return
	}

	end, err:= model.ParseStringToTime(sc.EndTime)
	if err!= nil {
		log.Println("string to time failed")
	}
	if time.Now().After(end) {
		appG.Response(http.StatusBadRequest, "活动已结束",  struct{}{})
		return
	}

	lockKey := fmt.Sprintf(gredis.COUPON_UID_LOCK, 2)
	drl := gredis.NewDisRedisLock(lockKey, uuid.NewString(), time.Second * 2)
	if !drl.TryLock() {
		appG.Response(http.StatusBadRequest,"失败", struct {}{})
		return
	}
	isExists = service.IsExistsCo(2, 2)
	// 检查是否已下单
	if isExists {
		appG.Response(http.StatusBadRequest,"请勿重复下单", struct {}{})
		return
	}
	insertId, err := service.AddCouponOrder(id)
	drl.AtomLockFree()
	if err != nil {
		appG.Response(http.StatusBadRequest, err.Error(),  struct{}{})
		return
	}
	appG.Response(http.StatusOK, "",  insertId)
}
