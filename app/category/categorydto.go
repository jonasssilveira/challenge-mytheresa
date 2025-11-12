package category

type CategoryDTO struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func ToCategoryDTO(category Category) CategoryDTO {
	return CategoryDTO{
		ID:   category.ID,
		Code: category.Code,
		Name: category.Name,
	}
}
