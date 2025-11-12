package catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mytheresa/go-hiring-challenge/app/category"
	"github.com/mytheresa/go-hiring-challenge/app/product"
	"github.com/mytheresa/go-hiring-challenge/app/product/mock"
	"github.com/mytheresa/go-hiring-challenge/app/variant"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetAllProducts(t *testing.T) {

	testProducts := []product.Product{
		{
			ID:         1,
			Code:       "PROD001",
			Price:      decimal.NewFromFloat(100.00),
			CategoryID: 1,
			Category:   category.Category{ID: 1, Code: "clothing", Name: "Clothing"},
			Variants:   []variant.Variant{},
		},
	}

	tests := []struct {
		name             string
		queryParams      string
		setupMock        func() mock.MockProduct
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:        "successfully returns paginated products",
			queryParams: "",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							if filter.Offset == 0 && filter.Limit == 10 {
								return product.ProductListResult{
									Products: testProducts,
									Total:    1,
								}, nil
							}
							return product.ProductListResult{}, errors.New("unexpected filter")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response CatalogResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.Total)
				assert.Len(t, response.Products, 1)
				assert.Equal(t, "PROD001", response.Products[0].Code)
			},
		},
		{
			name:        "returns products with category filter",
			queryParams: "?category=clothing",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							if filter.CategoryCode == "clothing" {
								return product.ProductListResult{
									Products: testProducts,
									Total:    1,
								}, nil
							}
							return product.ProductListResult{}, errors.New("unexpected filter")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "returns products with price filter",
			queryParams: "?priceLessThan=50",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							if filter.PriceLessThan != nil && *filter.PriceLessThan == 50.0 {
								return product.ProductListResult{
									Products: []product.Product{},
									Total:    0,
								}, nil
							}
							return product.ProductListResult{}, errors.New("unexpected filter")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "returns error for invalid priceLessThan parameter",
			queryParams:    "?priceLessThan=invalid",
			setupMock:      func() mock.MockProduct { return mock.MockProduct{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid priceLessThan")
			},
		},
		{
			name:        "enforces maximum limit of 100",
			queryParams: "?limit=200",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							if filter.Limit == 100 {
								return product.ProductListResult{
									Products: []product.Product{},
									Total:    0,
								}, nil
							}
							return product.ProductListResult{}, errors.New("expected limit to be 100")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "enforces minimum limit of 1",
			queryParams: "?limit=0",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							if filter.Limit == 1 {
								return product.ProductListResult{
									Products: []product.Product{},
									Total:    0,
								}, nil
							}
							return product.ProductListResult{}, errors.New("expected limit to be 1")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "handles repository error",
			queryParams: "",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							return product.ProductListResult{}, errors.New("database error")
						},
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
		{
			name:        "returns products with custom offset and limit",
			queryParams: "?offset=5&limit=20",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductsWithFilterMock: func(filter product.ProductFilter) (product.ProductListResult, error) {
							if filter.Offset == 5 && filter.Limit == 20 {
								return product.ProductListResult{
									Products: testProducts,
									Total:    25,
								}, nil
							}
							return product.ProductListResult{}, errors.New("unexpected filter")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			handler := NewCatalogHandler(mockRepo)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/catalog"+tt.queryParams, nil)
			c.Request = req

			handler.HandleGetAllProducts(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestHandleGetProductByCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testProduct := product.Product{
		ID:         1,
		Code:       "PROD001",
		Price:      decimal.NewFromFloat(100.00),
		CategoryID: 1,
		Category:   category.Category{ID: 1, Code: "clothing", Name: "Clothing"},
		Variants: []variant.Variant{
			{
				ID:        1,
				ProductID: 1,
				Name:      "Small",
				SKU:       "PROD001-S",
				Price:     decimal.Zero,
			},
		},
	}

	tests := []struct {
		name             string
		productCode      string
		setupMock        func() mock.MockProduct
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:        "successfully returns product by code",
			productCode: "PROD001",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductByCodeMock: func(code string) (product.Product, error) {
							if code == "PROD001" {
								return testProduct, nil
							}
							return product.Product{}, errors.New("not found")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response product.ProductDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "PROD001", response.Code)
				assert.Equal(t, "Clothing", response.Category.Name)
				assert.Len(t, response.Variants, 1)
				assert.Equal(t, "Small", response.Variants[0].Name)

				assert.True(t, response.Variants[0].Price.Equal(decimal.NewFromFloat(100.00)))
			},
		},
		{
			name:        "returns not found for non-existent product",
			productCode: "NONEXISTENT",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductByCodeMock: func(code string) (product.Product, error) {
							return product.Product{}, errors.New("not found")
						},
					},
				}
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Product not found", response["error"])
			},
		},
		{
			name:        "returns product with multiple variants",
			productCode: "PROD002",
			setupMock: func() mock.MockProduct {
				return mock.MockProduct{
					ProductRepositoryMock: mock.ProductRepositoryMock{
						GetProductByCodeMock: func(code string) (product.Product, error) {
							if code == "PROD002" {
								return product.Product{
									ID:         2,
									Code:       "PROD002",
									Price:      decimal.NewFromFloat(200.00),
									CategoryID: 1,
									Category:   category.Category{ID: 1, Code: "shoes", Name: "Shoes"},
									Variants: []variant.Variant{
										{
											ID:        2,
											ProductID: 2,
											Name:      "Size 8",
											SKU:       "PROD002-8",
											Price:     decimal.NewFromFloat(180.00),
										},
										{
											ID:        3,
											ProductID: 2,
											Name:      "Size 9",
											SKU:       "PROD002-9",
											Price:     decimal.Zero,
										},
									},
								}, nil
							}
							return product.Product{}, errors.New("not found")
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response product.ProductDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "PROD002", response.Code)
				assert.Len(t, response.Variants, 2)

				assert.True(t, response.Variants[0].Price.Equal(decimal.NewFromFloat(180.00)))

				assert.True(t, response.Variants[1].Price.Equal(decimal.NewFromFloat(200.00)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			handler := NewCatalogHandler(mockRepo)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "code", Value: tt.productCode}}
			req, _ := http.NewRequest("GET", "/catalog/"+tt.productCode, nil)
			c.Request = req

			handler.HandleGetProductByCode(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}
