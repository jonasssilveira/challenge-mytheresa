package product

import (
	"github.com/mytheresa/go-hiring-challenge/app/category"
	"github.com/mytheresa/go-hiring-challenge/app/variant"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID         uint              `gorm:"primaryKey"`
	Code       string            `gorm:"uniqueIndex;not null"`
	Price      decimal.Decimal   `gorm:"type:decimal(10,2);not null"`
	CategoryID uint              `gorm:"column:category_id"`
	Category   category.Category `gorm:"foreignKey:CategoryID;references:ID"`
	Variants   []variant.Variant `gorm:"foreignKey:ProductID;references:ID"`
}

func (p *Product) TableName() string {
	return "products"
}
