package category

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockCategoryRepository struct {
	GetAllCategoriesMock func() ([]Category, error)
	CreateCategoryMock   func(category Category) (Category, error)
}

func (m MockCategoryRepository) GetAllCategories() ([]Category, error) {
	return m.GetAllCategoriesMock()
}

func (m MockCategoryRepository) CreateCategory(cat Category) (Category, error) {
	return m.CreateCategoryMock(cat)
}

func TestHandleGetAllCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		setupMock        func() MockCategoryRepository
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "successfully returns all categories",
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					GetAllCategoriesMock: func() ([]Category, error) {
						return []Category{
							{ID: 1, Code: "clothing", Name: "Clothing"},
							{ID: 2, Code: "shoes", Name: "Shoes"},
							{ID: 3, Code: "accessories", Name: "Accessories"},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []CategoryDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 3)
				assert.Equal(t, "clothing", response[0].Code)
				assert.Equal(t, "Clothing", response[0].Name)
				assert.Equal(t, "shoes", response[1].Code)
				assert.Equal(t, "accessories", response[2].Code)
			},
		},
		{
			name: "handles empty category list",
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					GetAllCategoriesMock: func() ([]Category, error) {
						return []Category{}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []CategoryDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name: "handles repository error",
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					GetAllCategoriesMock: func() ([]Category, error) {
						return []Category{}, errors.New("database error")
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
			name: "returns categories with single item",
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					GetAllCategoriesMock: func() ([]Category, error) {
						return []Category{
							{ID: 1, Code: "electronics", Name: "Electronics"},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []CategoryDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
				assert.Equal(t, "electronics", response[0].Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			handler := NewCategoryHandler(mockRepo)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/categories", nil)
			c.Request = req

			handler.HandleGetAllCategories(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestHandleCreateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      interface{}
		setupMock        func() MockCategoryRepository
		expectedStatus   int
		validateResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "successfully creates a category",
			requestBody: CategoryDTO{
				Code: "electronics",
				Name: "Electronics",
			},
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					CreateCategoryMock: func(cat Category) (Category, error) {
						if cat.Code == "electronics" && cat.Name == "Electronics" {
							return Category{ID: 4, Code: "electronics", Name: "Electronics"}, nil
						}
						return Category{}, errors.New("unexpected category")
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response CategoryDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, uint(4), response.ID)
				assert.Equal(t, "electronics", response.Code)
				assert.Equal(t, "Electronics", response.Name)
			},
		},
		{
			name:           "returns error for invalid JSON",
			requestBody:    "invalid json",
			setupMock:      func() MockCategoryRepository { return MockCategoryRepository{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name: "returns error for missing code",
			requestBody: CategoryDTO{
				Name: "Electronics",
			},
			setupMock:      func() MockCategoryRepository { return MockCategoryRepository{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Code and Name are required", response["error"])
			},
		},
		{
			name: "returns error for missing name",
			requestBody: CategoryDTO{
				Code: "electronics",
			},
			setupMock:      func() MockCategoryRepository { return MockCategoryRepository{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Code and Name are required", response["error"])
			},
		},
		{
			name: "returns error for empty code",
			requestBody: CategoryDTO{
				Code: "",
				Name: "Electronics",
			},
			setupMock:      func() MockCategoryRepository { return MockCategoryRepository{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Code and Name are required", response["error"])
			},
		},
		{
			name: "handles repository error on create",
			requestBody: CategoryDTO{
				Code: "electronics",
				Name: "Electronics",
			},
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					CreateCategoryMock: func(cat Category) (Category, error) {
						return Category{}, errors.New("duplicate key")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "duplicate key", response["error"])
			},
		},
		{
			name: "successfully creates category with special characters",
			requestBody: CategoryDTO{
				Code: "home-and-garden",
				Name: "Home & Garden",
			},
			setupMock: func() MockCategoryRepository {
				return MockCategoryRepository{
					CreateCategoryMock: func(cat Category) (Category, error) {
						if cat.Code == "home-and-garden" && cat.Name == "Home & Garden" {
							return Category{ID: 5, Code: "home-and-garden", Name: "Home & Garden"}, nil
						}
						return Category{}, errors.New("unexpected category")
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response CategoryDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, uint(5), response.ID)
				assert.Equal(t, "home-and-garden", response.Code)
				assert.Equal(t, "Home & Garden", response.Name)
			},
		},
		{
			name: "returns error for empty name",
			requestBody: CategoryDTO{
				Code: "electronics",
				Name: "",
			},
			setupMock:      func() MockCategoryRepository { return MockCategoryRepository{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Code and Name are required", response["error"])
			},
		},
		{
			name: "handles whitespace only code",
			requestBody: CategoryDTO{
				Code: "   ",
				Name: "Electronics",
			},
			setupMock:      func() MockCategoryRepository { return MockCategoryRepository{} },
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Code and Name are required", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()
			handler := NewCategoryHandler(mockRepo)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("POST", "/categories", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.HandleCreateCategory(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}
