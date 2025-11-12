package variant

import "github.com/shopspring/decimal"

type VariantDTO struct {
	ID        uint            `json:"id"`
	ProductID uint            `json:"product_id"`
	Name      string          `json:"name"`
	SKU       string          `json:"sku"`
	Price     decimal.Decimal `json:"price"`
}

func ToVariantDTO(variant Variant, productPrice decimal.Decimal) VariantDTO {
	return VariantDTO{
		ID:        variant.ID,
		ProductID: variant.ProductID,
		Name:      variant.Name,
		SKU:       variant.SKU,
		Price:     setPrice(variant.Price, productPrice),
	}
}

func ToVariantsDTO(variants []Variant, productPrice decimal.Decimal) []VariantDTO {
	variantSlice := make([]VariantDTO, 0, len(variants))
	for _, v := range variants {
		variantSlice = append(variantSlice, ToVariantDTO(v, productPrice))
	}
	return variantSlice
}

func setPrice(variantPrice decimal.Decimal, productPrice decimal.Decimal) decimal.Decimal {
	if variantPrice.Equal(decimal.Zero) {
		return productPrice
	}
	return variantPrice
}
