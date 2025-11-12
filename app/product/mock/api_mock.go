package mock

import "github.com/mytheresa/go-hiring-challenge/app/product"

type ProductRepositoryMock struct {
	GetAllProductsMock        func() ([]product.Product, error)
	GetProductByCodeMock      func(code string) (product.Product, error)
	GetProductsWithFilterMock func(filter product.ProductFilter) (product.ProductListResult, error)
}

type MockProduct struct {
	ProductRepositoryMock
}

func (mock MockProduct) GetAllProducts() ([]product.Product, error) {
	return mock.GetAllProductsMock()
}

func (mock MockProduct) GetProductByCode(code string) (product.Product, error) {
	return mock.GetProductByCodeMock(code)
}

func (mock MockProduct) GetProductsWithFilter(filter product.ProductFilter) (product.ProductListResult, error) {
	return mock.GetProductsWithFilterMock(filter)
}
