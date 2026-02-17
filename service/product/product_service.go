package product

import (
	productEntity "magento.GO/model/entity/product"
	productRepository "magento.GO/model/repository/product"
)

type ProductInput struct {
	AttributeSetID  uint16
	TypeID          string
	SKU             string
	HasOptions      uint16
	RequiredOptions uint16
}

type ProductService struct {
	repo *productRepository.ProductRepository
}

func NewProductService(repo *productRepository.ProductRepository) *ProductService {
	return &ProductService{repo}
}

func (s *ProductService) ListProducts() ([]productEntity.Product, error) {
	return s.repo.FindAll()
}

func (s *ProductService) GetProduct(id uint) (*productEntity.Product, error) {
	return s.repo.FindByID(id)
}

func (s *ProductService) CreateProduct(input *ProductInput) error {
	prod := &productEntity.Product{
		AttributeSetID:  input.AttributeSetID,
		TypeID:          input.TypeID,
		SKU:             input.SKU,
		HasOptions:      input.HasOptions,
		RequiredOptions: input.RequiredOptions,
	}
	return s.repo.Create(prod)
}

func (s *ProductService) UpdateProduct(id uint, input *ProductInput) error {
	prod, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	prod.AttributeSetID = input.AttributeSetID
	prod.TypeID = input.TypeID
	prod.SKU = input.SKU
	prod.HasOptions = input.HasOptions
	prod.RequiredOptions = input.RequiredOptions
	return s.repo.Update(prod)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.repo.Delete(id)
}
