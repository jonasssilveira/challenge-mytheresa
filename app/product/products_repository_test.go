package product

import (
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/category"
	"github.com/mytheresa/go-hiring-challenge/app/variant"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&Product{}, &category.Category{}, &variant.Variant{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func seedTestData(t *testing.T, db *gorm.DB) {

	categories := []category.Category{
		{ID: 1, Code: "clothing", Name: "Clothing"},
		{ID: 2, Code: "shoes", Name: "Shoes"},
		{ID: 3, Code: "accessories", Name: "Accessories"},
	}

	for _, cat := range categories {
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("Failed to seed category: %v", err)
		}
	}

	products := []Product{
		{
			ID:         1,
			Code:       "PROD001",
			Price:      decimal.NewFromFloat(100.00),
			CategoryID: 1,
		},
		{
			ID:         2,
			Code:       "PROD002",
			Price:      decimal.NewFromFloat(50.00),
			CategoryID: 2,
		},
		{
			ID:         3,
			Code:       "PROD003",
			Price:      decimal.NewFromFloat(75.00),
			CategoryID: 3,
		},
		{
			ID:         4,
			Code:       "PROD004",
			Price:      decimal.NewFromFloat(150.00),
			CategoryID: 1,
		},
	}

	for _, prod := range products {
		if err := db.Create(&prod).Error; err != nil {
			t.Fatalf("Failed to seed product: %v", err)
		}
	}

	variants := []variant.Variant{
		{
			ID:        1,
			ProductID: 1,
			Name:      "Small",
			SKU:       "PROD001-S",
			Price:     decimal.Zero,
		},
		{
			ID:        2,
			ProductID: 1,
			Name:      "Medium",
			SKU:       "PROD001-M",
			Price:     decimal.NewFromFloat(110.00),
		},
		{
			ID:        3,
			ProductID: 2,
			Name:      "Size 8",
			SKU:       "PROD002-8",
			Price:     decimal.Zero,
		},
	}

	for _, v := range variants {
		if err := db.Create(&v).Error; err != nil {
			t.Fatalf("Failed to seed variant: %v", err)
		}
	}
}

func TestGetAllProducts(t *testing.T) {
	tests := []struct {
		name           string
		seedData       bool
		expectedCount  int
		expectError    bool
		validateResult func(t *testing.T, products []Product)
	}{
		{
			name:          "successfully returns all products",
			seedData:      true,
			expectedCount: 4,
			expectError:   false,
			validateResult: func(t *testing.T, products []Product) {
				assert.Len(t, products, 4)

				prod := products[0]
				assert.NotEmpty(t, prod.Category.Name)
				assert.NotNil(t, prod.Variants)
			},
		},
		{
			name:          "returns empty list when no products exist",
			seedData:      false,
			expectedCount: 0,
			expectError:   false,
			validateResult: func(t *testing.T, products []Product) {
				assert.Len(t, products, 0)
			},
		},
		{
			name:          "preloads category relationships",
			seedData:      true,
			expectedCount: 4,
			expectError:   false,
			validateResult: func(t *testing.T, products []Product) {
				for _, prod := range products {
					assert.NotZero(t, prod.Category.ID, "Category should be loaded for product %s", prod.Code)
					assert.NotEmpty(t, prod.Category.Name, "Category name should be loaded for product %s", prod.Code)
				}
			},
		},
		{
			name:          "preloads variants relationships",
			seedData:      true,
			expectedCount: 4,
			expectError:   false,
			validateResult: func(t *testing.T, products []Product) {

				prod001 := findProductByCode(products, "PROD001")
				assert.NotNil(t, prod001)
				assert.Len(t, prod001.Variants, 2, "PROD001 should have 2 variants")

				prod002 := findProductByCode(products, "PROD002")
				assert.NotNil(t, prod002)
				assert.Len(t, prod002.Variants, 1, "PROD002 should have 1 variant")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tt.seedData {
				seedTestData(t, db)
			}

			repo := NewProductsRepository(db)
			products, err := repo.GetAllProducts()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, products, tt.expectedCount)

				if tt.validateResult != nil {
					tt.validateResult(t, products)
				}
			}
		})
	}
}

func TestGetProductByCode(t *testing.T) {
	tests := []struct {
		name           string
		productCode    string
		seedData       bool
		expectError    bool
		validateResult func(t *testing.T, product Product)
	}{
		{
			name:        "successfully returns product by code",
			productCode: "PROD001",
			seedData:    true,
			expectError: false,
			validateResult: func(t *testing.T, product Product) {
				assert.Equal(t, "PROD001", product.Code)
				assert.True(t, product.Price.Equal(decimal.NewFromFloat(100.00)))
				assert.NotZero(t, product.ID)
			},
		},
		{
			name:        "preloads category for product",
			productCode: "PROD001",
			seedData:    true,
			expectError: false,
			validateResult: func(t *testing.T, product Product) {
				assert.Equal(t, "Clothing", product.Category.Name)
				assert.Equal(t, "clothing", product.Category.Code)
				assert.Equal(t, uint(1), product.CategoryID)
			},
		},
		{
			name:        "preloads variants for product",
			productCode: "PROD001",
			seedData:    true,
			expectError: false,
			validateResult: func(t *testing.T, product Product) {
				assert.Len(t, product.Variants, 2)
				assert.Equal(t, "Small", product.Variants[0].Name)
				assert.Equal(t, "Medium", product.Variants[1].Name)
			},
		},
		{
			name:        "returns error for non-existent product",
			productCode: "NONEXISTENT",
			seedData:    true,
			expectError: true,
		},
		{
			name:        "returns error when database is empty",
			productCode: "PROD001",
			seedData:    false,
			expectError: true,
		},
		{
			name:        "handles product with no variants",
			productCode: "PROD003",
			seedData:    true,
			expectError: false,
			validateResult: func(t *testing.T, product Product) {
				assert.Equal(t, "PROD003", product.Code)
				assert.Len(t, product.Variants, 0)
				assert.NotNil(t, product.Variants)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tt.seedData {
				seedTestData(t, db)
			}

			repo := NewProductsRepository(db)
			product, err := repo.GetProductByCode(tt.productCode)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, product)
				}
			}
		})
	}
}

func TestGetProductsWithFilter(t *testing.T) {
	tests := []struct {
		name           string
		filter         ProductFilter
		seedData       bool
		expectedTotal  int64
		expectedCount  int
		expectError    bool
		validateResult func(t *testing.T, result ProductListResult)
	}{
		{
			name: "returns all products with default filter",
			filter: ProductFilter{
				Offset: 0,
				Limit:  10,
			},
			seedData:      true,
			expectedTotal: 4,
			expectedCount: 4,
			expectError:   false,
		},
		{
			name: "filters by category code",
			filter: ProductFilter{
				CategoryCode: "clothing",
				Offset:       0,
				Limit:        10,
			},
			seedData:      true,
			expectedTotal: 2,
			expectedCount: 2,
			expectError:   false,
			validateResult: func(t *testing.T, result ProductListResult) {
				assert.Equal(t, int64(2), result.Total)
				for _, prod := range result.Products {
					assert.Equal(t, "clothing", prod.Category.Code)
				}
			},
		},
		{
			name: "filters by price less than",
			filter: ProductFilter{
				PriceLessThan: 100.00,
				Offset:        0,
				Limit:         10,
			},
			seedData:      true,
			expectedTotal: 2,
			expectedCount: 2,
			expectError:   false,
			validateResult: func(t *testing.T, result ProductListResult) {
				for _, prod := range result.Products {
					assert.True(t, prod.Price.LessThan(decimal.NewFromFloat(100.00)),
						"Product %s price should be less than 100.00", prod.Code)
				}
			},
		},
		{
			name: "combines category and price filters",
			filter: ProductFilter{
				CategoryCode:  "clothing",
				PriceLessThan: 120.00,
				Offset:        0,
				Limit:         10,
			},
			seedData:      true,
			expectedTotal: 1,
			expectedCount: 1,
			expectError:   false,
			validateResult: func(t *testing.T, result ProductListResult) {
				assert.Equal(t, int64(1), result.Total)
				assert.Equal(t, "PROD001", result.Products[0].Code)
			},
		},
		{
			name: "applies pagination with offset",
			filter: ProductFilter{
				Offset: 2,
				Limit:  2,
			},
			seedData:      true,
			expectedTotal: 4,
			expectedCount: 2,
			expectError:   false,
			validateResult: func(t *testing.T, result ProductListResult) {
				assert.Equal(t, int64(4), result.Total)
				assert.Len(t, result.Products, 2)
			},
		},
		{
			name: "applies pagination with limit",
			filter: ProductFilter{
				Offset: 0,
				Limit:  2,
			},
			seedData:      true,
			expectedTotal: 4,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "returns empty result for non-matching category",
			filter: ProductFilter{
				CategoryCode: "nonexistent",
				Offset:       0,
				Limit:        10,
			},
			seedData:      true,
			expectedTotal: 0,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "returns empty result for very low price filter",
			filter: ProductFilter{
				PriceLessThan: 1.00,
				Offset:        0,
				Limit:         10,
			},
			seedData:      true,
			expectedTotal: 0,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "handles offset beyond total count",
			filter: ProductFilter{
				Offset: 10,
				Limit:  10,
			},
			seedData:      true,
			expectedTotal: 4,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "preloads relationships with filters",
			filter: ProductFilter{
				CategoryCode: "clothing",
				Offset:       0,
				Limit:        10,
			},
			seedData:      true,
			expectedTotal: 2,
			expectedCount: 2,
			expectError:   false,
			validateResult: func(t *testing.T, result ProductListResult) {
				for _, prod := range result.Products {
					assert.NotEmpty(t, prod.Category.Name, "Category should be loaded")
					assert.NotNil(t, prod.Variants, "Variants should be loaded")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			if tt.seedData {
				seedTestData(t, db)
			}

			repo := NewProductsRepository(db)
			result, err := repo.GetProductsWithFilter(tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTotal, result.Total)
				assert.Len(t, result.Products, tt.expectedCount)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func findProductByCode(products []Product, code string) *Product {
	for _, p := range products {
		if p.Code == code {
			return &p
		}
	}
	return nil
}
