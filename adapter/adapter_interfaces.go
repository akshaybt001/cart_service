package adapter

import "github.com/akshaybt001/cart_service/entities"

type AdapterInterface interface {
	CreateCart(userId uint) error
	AddToCart(req entities.CartItems,userId uint)error
	GetAllFromCart(userId uint)([]entities.CartItems,error)
	RemoveFromCart(req entities.CartItems,userId uint)error
	IsEmpty(req entities.CartItems,userId uint)bool
	TruncateCart(userId int)error
}