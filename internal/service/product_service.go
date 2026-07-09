package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
)

type ProductService struct {
	productRepo *postgres.ProductRepository
}

func NewProductService(productRepo *postgres.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) CreateProduct(name, description, unit string, price float64) (*entity.Product, error) {
	if name == "" {
		return nil, errors.New("nama produk wajib diisi")
	}
	if price < 0 {
		return nil, errors.New("harga tidak boleh negatif")
	}
	if unit == "" {
		unit = "pcs"
	}

	product := &entity.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Unit:        unit,
		IsActive:    true,
	}

	return s.productRepo.CreateProduct(product)
}

func (s *ProductService) GetAllProducts(isActive *bool) ([]entity.Product, error) {
	return s.productRepo.GetAllProducts(isActive)
}

func (s *ProductService) GetProductByID(productID string) (*entity.Product, error) {
	if productID == "" {
		return nil, errors.New("ID produk wajib diisi")
	}
	return s.productRepo.GetProductByID(productID)
}

func (s *ProductService) UpdateProduct(productID string, name, description, unit *string, price *float64, isActive *bool) error {
	if productID == "" {
		return errors.New("ID produk wajib diisi")
	}

	_, err := s.productRepo.GetProductByID(productID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if name != nil {
		if *name == "" {
			return errors.New("nama produk tidak boleh kosong")
		}
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}
	if unit != nil {
		updates["unit"] = *unit
	}
	if price != nil {
		if *price < 0 {
			return errors.New("harga tidak boleh negatif")
		}
		updates["price"] = *price
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}

	if len(updates) == 0 {
		return errors.New("tidak ada data yang diupdate")
	}

	return s.productRepo.UpdateProduct(productID, updates)
}

func (s *ProductService) DeleteProduct(productID string) error {
	if productID == "" {
		return errors.New("ID produk wajib diisi")
	}
	_, err := s.productRepo.GetProductByID(productID)
	if err != nil {
		return err
	}
	return s.productRepo.DeleteProduct(productID)
}
