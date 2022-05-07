package model

import "time"

type Coupon struct {
	Id uint64 `json:"id"`
	Type uint8 `json:"type"`
	ReplacePrice float32 `json:"replace_price"`
	Price float32 `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	ShopId uint8 `json:"shop_id"`
	PayType uint8 `json:"pay_type"`
}
