package model

import (
	"database/sql"
	"time"
)

type CouponOrder struct {

	Id uint64 `json:"id"`
	Status uint8 `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Cid uint64 `json:"cid"`
	Uid uint64 `json:"uid"`
}

func AddCouponOrder (co *CouponOrder) (sql.Result, error) {
	sql := "insert into coupon_order (`c_id`, `uid`) value(?, ?)"
	return db.Exec(sql, co.Cid, co.Uid)
}

func GetCOByUidAndCid (cid, uid int)  *sql.Row {
	//var co CouponOrder
	sqlStr := "select * from coupon_order where c_id = ? and uid = ?"
	//err := db.QueryRow(sqlStr, cid, uid).Scan(&co.Id, &co.Status, &co.CreatedAt, &co.Cid, &co.Uid )
	return db.QueryRow(sqlStr, cid, uid)
}

