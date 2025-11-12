package product

import (
	"github.com/mytheresa/go-hiring-challenge/app/category"
	"github.com/mytheresa/go-hiring-challenge/app/variant"
)

type ProductDTO struct {
	ID       uint                 `json:"id"`
	Code     string               `json:"code"`
	Price    float64              `json:"price"`
	Category category.CategoryDTO `json:"category"`
	Variants []variant.VariantDTO `json:"variants"`
}
