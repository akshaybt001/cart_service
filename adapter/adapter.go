package adapter

import (
	"fmt"

	"github.com/akshaybt001/cart_service/entities"
	"gorm.io/gorm"
)

type CartAdapter struct {
	DB *gorm.DB
}

func NewCartAdapter(db *gorm.DB) *CartAdapter {
	return &CartAdapter{
		DB: db,
	}
}

func (cart *CartAdapter) CreateCart(userId uint) error {
	query := "INSERT INTO carts (user_id) VALUES($1)"
	if err := cart.DB.Exec(query, userId).Error; err != nil {
		return err
	}
	return nil
}

func (cart *CartAdapter) AddToCart(req entities.CartItems, userId uint) error {
	tx := cart.DB.Begin()

	var cartId int
	var current entities.CartItems

	quaryId := "SELECT id FROM carts WHERE user_id = ?"

	if err := tx.Raw(quaryId, userId).Scan(&cartId).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("cart not found")
	}
	queryCurrent := "SELECT * FROM cart_items WHERE cart_id = $1 AND product_id = $2"
	if err := tx.Raw(queryCurrent, cartId, req.ProductId).Scan(&current).Error; err != nil {
		tx.Rollback()
		return err
	}
	var res entities.CartItems
	if current.ProductId == 0 {
		insertQuery := "INSERT INTO cart_items(cart_id ,product_id ,quantity,total) VALUES ($1,$2,$3,0) RETURNING id ,product_id,cart_id"
		if err := tx.Raw(insertQuery, cartId, req.ProductId, req.Quantity).Scan(&res).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		updateQuery := "UPDATE cart_items SET  quantity = quantity + $1 WHERE cart_id = $2 AND  product_id = $3"
		if err := tx.Exec(updateQuery, req.Quantity, cartId, req.ProductId).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	updateTotal := "UPDATE cart_items SET total = total + $1 WHERE cart_id = $2 AND product_id = $3"
	if err := tx.Exec(updateTotal, (req.Total * float64(req.Quantity)), cartId, req.ProductId).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (cart *CartAdapter) GetAllFromCart(userId uint) ([]entities.CartItems, error) {
	var res []entities.CartItems
	var cartId uint

	quary := "SELECT id FROM carts WHERE user_id = ?"

	if err := cart.DB.Raw(quary, userId).Scan(&cartId).Error; err != nil {
		return nil, err
	}

	quaryItems := "SELECT * FROM cart_items WHERE cart_id = ?"
	if err := cart.DB.Raw(quaryItems, cartId).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (cart *CartAdapter) RemoveFromCart(req entities.CartItems, userId uint) error {
	tx := cart.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var cartId uint

	query := "SELECT id FROM carts WHERE user_id = ?"
	if err := cart.DB.Raw(query, userId).Scan(&cartId).Error; err != nil {
		tx.Rollback()
		return err
	}

	var current entities.CartItems
	queryItems := "SELECT * FROM carts WHERE user_id = $1 AND product_id = $2"
	if err := cart.DB.Raw(queryItems, cartId, req.ProductId).Scan(&current).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("there is no such product in cart")
	}
	if current.ProductId == 0 {
		return fmt.Errorf("there is no such product in cart")
	}
	if current.Quantity <= 0 {
		return fmt.Errorf("there is no such product in cart")
	}

	queryUpdate := "UPDATE cart_items SET quantity = quantity - 1 WHERE cart_id = $1 AND product_id = $2"
	if err := cart.DB.Exec(queryUpdate, cartId, req.ProductId).Error; err != nil {
		tx.Rollback()
		return err
	}

	var quantity int

	queryUpdateTotal := "UPDATE cart_items SET total = total - $1 WHERE cart_id = $2 AND product_id =  $3 RETURNING quantity"
	if err := cart.DB.Raw(queryUpdateTotal, req.Total, cartId, req.ProductId).Scan(&quantity).Error; err != nil {
		tx.Rollback()
		return err
	}
	if quantity == 0 {
		queryDelete := "DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2"
		if err := cart.DB.Exec(queryDelete, cartId, req.ProductId).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error while commiting")
	}
	return nil
}

func (cart *CartAdapter) IsEmpty(req entities.CartItems, userId uint) bool {
	var cartId uint

	queryId := "SELECT id FROM carts WHERE user_id = ?"
	if err := cart.DB.Raw(queryId, userId).Scan(&cartId).Error; err != nil {
		return true
	}
	var cartItem entities.CartItems
	queryCheck := "SELECT * FROM  cart_items WHERE cart_id = ?"
	if err := cart.DB.Raw(queryCheck, cartId).Scan(&cartItem).Error; err != nil {
		return true
	}
	if cartItem.CartID == 0 {
		return true
	}
	return false
}


func (cart *CartAdapter) TruncateCart(userId int) error {
	var cartId int
	queryId:="SELECT id FROM carts WHERE user_id = ?"
	if err:=cart.DB.Raw(queryId,userId).Scan(&cartId).Error;err!=nil{
		return err
	}
	query:="DELETE FROM cart_items WHERE cart_id = ?"
	tx:=cart.DB.Begin()
	if err:=tx.Exec(query,cartId).Error;err!=nil{
		tx.Rollback()
		return err
	}
	if err:=tx.Commit().Error;err!=nil{
		return err
	}
	return nil
}
