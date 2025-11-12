package category

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mytheresa/go-hiring-challenge/app/api"
)

type CategoryHandler struct {
	repo CategoryRepositoryInterface
}

func NewCategoryHandler(r CategoryRepositoryInterface) *CategoryHandler {
	return &CategoryHandler{
		repo: r,
	}
}

func (h *CategoryHandler) HandleGetAllCategories(c *gin.Context) {
	w := c.Writer

	categories, err := h.repo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	categoryDTOs := make([]CategoryDTO, len(categories))
	for i, cat := range categories {
		categoryDTOs[i] = ToCategoryDTO(cat)
	}

	api.OKResponse(w, categoryDTOs)
}

func (h *CategoryHandler) HandleCreateCategory(c *gin.Context) {
	w := c.Writer
	var categoryDTO CategoryDTO

	if err := c.ShouldBindJSON(&categoryDTO); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	categoryDTO.Code = strings.TrimSpace(categoryDTO.Code)
	categoryDTO.Name = strings.TrimSpace(categoryDTO.Name)

	if categoryDTO.Code == "" || categoryDTO.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Code and Name are required")
		return
	}

	category := Category{
		Code: categoryDTO.Code,
		Name: categoryDTO.Name,
	}

	createdCategory, err := h.repo.CreateCategory(category)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.OKResponse(w, ToCategoryDTO(createdCategory))
}
