package catalog

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/category"
	"github.com/mytheresa/go-hiring-challenge/app/product"
	"github.com/mytheresa/go-hiring-challenge/app/variant"
)

type CatalogResponse struct {
	Products []product.ProductDTO `json:"products"`
	Total    int64                `json:"total"`
}

type CatalogHandler struct {
	repo product.ProductRepository
}

func NewCatalogHandler(r product.ProductRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGetAllProducts(c *gin.Context) {
	w := c.Writer

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	categoryCode := c.Query("category")
	priceLessThanStr := c.Query("priceLessThan")

	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	filter := product.ProductFilter{
		CategoryCode: categoryCode,
		Offset:       offset,
		Limit:        limit,
	}

	if priceLessThanStr != "" {
		priceLessThan, err := strconv.ParseFloat(priceLessThanStr, 64)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid priceLessThan parameter")
			return
		}
		filter.PriceLessThan = &priceLessThan
	}

	result, err := h.repo.GetProductsWithFilter(filter)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	products := make([]product.ProductDTO, len(result.Products))
	for i, p := range result.Products {
		products[i] = product.ProductDTO{
			ID:       p.ID,
			Variants: variant.ToVariantsDTO(p.Variants, p.Price),
			Code:     p.Code,
			Category: category.ToCategoryDTO(p.Category),
			Price:    p.Price.InexactFloat64(),
		}
	}

	response := CatalogResponse{
		Products: products,
		Total:    result.Total,
	}

	api.OKResponse(w, response)
}

func (h *CatalogHandler) HandleGetProductByCode(c *gin.Context) {
	w := c.Writer
	code := c.Param("code")

	res, err := h.repo.GetProductByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	productDTO := product.ProductDTO{
		ID:       res.ID,
		Variants: variant.ToVariantsDTO(res.Variants, res.Price),
		Code:     res.Code,
		Category: category.ToCategoryDTO(res.Category),
		Price:    res.Price.InexactFloat64(),
	}

	api.OKResponse(w, productDTO)
}
