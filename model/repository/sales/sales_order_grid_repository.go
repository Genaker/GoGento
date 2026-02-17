package sales

import (
	"gorm.io/gorm"
	salesEntity "magento.GO/model/entity/sales"
)

type SalesOrderGridRepository struct {
	db *gorm.DB
}

func NewSalesOrderGridRepository(db *gorm.DB) *SalesOrderGridRepository {
	return &SalesOrderGridRepository{db}
}

func (r *SalesOrderGridRepository) FindAll() ([]salesEntity.SalesOrderGrid, error) {
	var orders []salesEntity.SalesOrderGrid
	err := r.db.Find(&orders).Error
	return orders, err
}

func (r *SalesOrderGridRepository) FindByID(id uint) (*salesEntity.SalesOrderGrid, error) {
	var order salesEntity.SalesOrderGrid
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *SalesOrderGridRepository) Create(order *salesEntity.SalesOrderGrid) error {
	return r.db.Create(order).Error
}

func (r *SalesOrderGridRepository) Update(order *salesEntity.SalesOrderGrid) error {
	return r.db.Save(order).Error
}

func (r *SalesOrderGridRepository) Delete(id uint) error {
	return r.db.Delete(&salesEntity.SalesOrderGrid{}, id).Error
}
