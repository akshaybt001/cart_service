package initializer

import (
	"github.com/akshaybt001/cart_service/adapter"
	"github.com/akshaybt001/cart_service/service"
	"gorm.io/gorm"
)

func Initializer(db *gorm.DB) *service.CartService{
	adapter:=adapter.NewCartAdapter(db)
	service:=service.NewCartService(adapter)

	return service
}