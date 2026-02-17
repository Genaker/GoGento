package product

import "time"

type ProductDatetime struct {
	ValueID     uint      `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16    `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16    `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint      `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       time.Time `gorm:"column:value"`
}

func (ProductDatetime) TableName() string {
	return "catalog_product_entity_datetime"
}

/* Usage Examples:
1. Create:
   attr := &ProductDatetime{AttributeID: 55, StoreID: 1, EntityID: 123, Value: time.Now()}
   db.Create(attr)
2. Read:
   var attr ProductDatetime
   db.First(&attr, valueID)
3. Update:
   db.Model(&attr).Update("Value", time.Now())
4. Delete:
   db.Delete(&attr)
*/
