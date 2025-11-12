package product

import (
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db gorm.DB
}

type ProductFilter struct {
	CategoryCode  string
	PriceLessThan *float64
	Offset        int
	Limit         int
}

type ProductListResult struct {
	Products []Product
	Total    int64
}

type ProductRepository interface {
	GetAllProducts() ([]Product, error)
	GetProductByCode(code string) (Product, error)
	GetProductsWithFilter(filter ProductFilter) (ProductListResult, error)
}

func NewProductsRepository(db *gorm.DB) ProductsRepository {
	return ProductsRepository{
		db: *db,
	}
}

func (r ProductsRepository) GetProductByCode(code string) (Product, error) {
	var product Product
	if err := r.db.Preload("Category").Preload("Variants").Where("code = ?", code).First(&product).Error; err != nil {
		return Product{}, err
	}
	return product, nil
}

func (r ProductsRepository) GetAllProducts() ([]Product, error) {
	var products []Product
	if err := r.db.Preload("Category").Preload("Variants").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r ProductsRepository) GetProductsWithFilter(filter ProductFilter) (ProductListResult, error) {
	var products []Product
	var total int64

	query := r.db.Model(&Product{})

	if filter.CategoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", filter.CategoryCode)
	}

	if filter.PriceLessThan != nil {
		query = query.Where("products.price < ?", *filter.PriceLessThan)
	}

	if err := query.Count(&total).Error; err != nil {
		return ProductListResult{}, err
	}

	query = query.Offset(filter.Offset).Limit(filter.Limit)

	if err := query.Preload("Category").Preload("Variants").Find(&products).Error; err != nil {
		return ProductListResult{}, err
	}

	return ProductListResult{
		Products: products,
		Total:    total,
	}, nil
}
