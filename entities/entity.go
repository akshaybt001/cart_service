package entities

type Cart struct{
	Id uint
	UserId uint
}

type CartItems struct{
	Id uint
	CartID uint
	ProductId uint
	Quantity int
	Total float64
}