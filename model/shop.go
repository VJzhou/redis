package model

import (
	"encoding/json"
	"time"
)

type Shop struct {
	Id int64 `json:"id"`
	ShopName string `json:"shop_name"`
	ShopDesc string `json:"shop_desc"`
	ShopAddr string `json:"shop_addr"`
	ShopPhone string `json:"shop_phone"`
	CreatedAt time.Time `json:"created_at"`
}
type RedisShop struct {
	Shop
	Expire int64
}

func (s RedisShop) MarshalBinary () ([]byte, error){
	return json.Marshal(s)
}

func (s Shop) MarshalBinary () ([]byte, error){
	return json.Marshal(s)
}

