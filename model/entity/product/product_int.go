package product

type ProductInt struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16 `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint   `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       int    `gorm:"column:value"`
}

func (ProductInt) TableName() string {
	return "catalog_product_entity_int"
}

/* Usage Examples:
1. Create:
   attr := &ProductInt{AttributeID: 99, StoreID: 1, EntityID: 123, Value: 1}
   db.Create(attr)
2. Read:
   var attr ProductInt
   db.First(&attr, valueID)
3. Update:
   db.Model(&attr).Update("Value", 2)
4. Delete:
   db.Delete(&attr)
*/
