package product

type ProductText struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16 `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint   `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       string `gorm:"column:value;type:text"`
}

func (ProductText) TableName() string {
	return "catalog_product_entity_text"
}

/* Usage Examples:
1. Create:
   attr := &ProductText{AttributeID: 88, StoreID: 1, EntityID: 123, Value: "Long description"}
   db.Create(attr)
2. Read:
   var attr ProductText
   db.First(&attr, valueID)
3. Update:
   db.Model(&attr).Update("Value", "Updated description")
4. Delete:
   db.Delete(&attr)
*/
