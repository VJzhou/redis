package service

import (
	"errors"
	"redis/model"
)

func AddCouponOrder (cid int) (int64, error) {
	db := model.NewDB()

	tx, err := db.Begin()
	var co model.CouponOrder
	co.Cid = uint64(cid)
	co.Uid = 2

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

