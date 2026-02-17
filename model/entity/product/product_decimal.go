package product

type ProductDecimal struct {
	ValueID     uint    `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16  `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16  `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint    `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       float64 `gorm:"column:value"`
}

func (ProductDecimal) TableName() string {
	return "catalog_product_entity_decimal"
}

/* Usage Examples:
1. Create:
   attr := &ProductDecimal{AttributeID: 77, StoreID: 1, EntityID: 123, Value: 99.99}
   db.Create(attr)
2. Read:
   var attr ProductDecimal
   db.First(&attr, valueID)
3. Update:
   db.Model(&attr).Update("Value", 100.00)
4. Delete:
   db.Delete(&attr)
*/
