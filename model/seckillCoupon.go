package model

import (
	"database/sql"
	"errors"
)

type SeckillCoupon struct {
	Id uint64 `json:"id"`
	Cid uint64 `json:"cid"`
	BeginTime string `json:"begin_time"`
	EndTime string `json:"end_time"`
	Stock uint8 `json:"stock"`
	CreatedAt string `json:"created_at"`
}

func GetSCRow (cid int) (*SeckillCoupon, error)  {
	row := db.QueryRow("select * from seckill_coupon where c_id = ?", cid)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var sc SeckillCoupon
	if err := row.Scan(&sc.Id, &sc.Cid, &sc.BeginTime, &sc.EndTime, &sc.Stock, &sc.CreatedAt); err != nil {
		return nil, errors.New("scan failed")
	}
	return &sc, nil
}

func DecrByStep (Cid , step int) (sql.Result, error) {
	sqlstr := "update seckill_coupon set stock = stock  - ? where c_id = ? and stock > 0"
	return  db.Exec(sqlstr, step, Cid)
}
