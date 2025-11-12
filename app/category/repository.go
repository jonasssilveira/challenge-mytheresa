package category

import (
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db gorm.DB
}

type CategoryRepositoryInterface interface {
	GetAllCategories() ([]Category, error)
	CreateCategory(category Category) (Category, error)
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return CategoryRepository{
		db: *db,
	}
}

func (r CategoryRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r CategoryRepository) CreateCategory(category Category) (Category, error) {
	if err := r.db.Create(&category).Error; err != nil {
		return Category{}, err
	}
	return category, nil
}
