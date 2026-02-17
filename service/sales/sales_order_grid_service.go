package sales

import (
	entity "magento.GO/model/entity/sales"
	repository "magento.GO/model/repository/sales"
)

type SalesOrderGridService struct {
	repo *repository.SalesOrderGridRepository
}

func NewSalesOrderGridService(repo *repository.SalesOrderGridRepository) *SalesOrderGridService {
	return &SalesOrderGridService{repo}
}

func (s *SalesOrderGridService) ListOrders() ([]entity.SalesOrderGrid, error) {
	return s.repo.FindAll()
}

func (s *SalesOrderGridService) GetOrder(id uint) (*entity.SalesOrderGrid, error) {
	return s.repo.FindByID(id)
}

func (s *SalesOrderGridService) CreateOrder(order *entity.SalesOrderGrid) error {
	return s.repo.Create(order)
}

func (s *SalesOrderGridService) UpdateOrder(order *entity.SalesOrderGrid) error {
	return s.repo.Update(order)
}

func (s *SalesOrderGridService) DeleteOrder(id uint) error {
	return s.repo.Delete(id)
}
