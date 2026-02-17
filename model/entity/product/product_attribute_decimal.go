package product

type ProductAttributeDecimal struct {
	ValueID     uint    `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16  `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16  `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint    `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       float64 `gorm:"column:value;type:decimal(20,6)"`
}

// TableName specifies the table name
func (ProductAttributeDecimal) TableName() string {
	return "catalog_product_entity_decimal"
}

/* Usage Examples:

1. Create:
   ```go
   attrDecimal := &ProductAttributeDecimal{
       AttributeID: 1,
       StoreID: 1,
       EntityID: 1,
       Value: 123.456,
   }
   db.Create(attrDecimal)
   ```

2. Read:
   ```go
   var attrDecimal ProductAttributeDecimal
   db.First(&attrDecimal, valueID)
   ```

3. Update:
   ```go
   db.Model(&attrDecimal).Update("Value", 789.012)
   ```

4. Delete:
   ```go
   db.Delete(&attrDecimal)
   ```
*/
