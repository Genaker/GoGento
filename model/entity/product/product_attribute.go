package product

type ProductAttributeInt struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16 `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint   `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       int    `gorm:"column:value"`
}

// TableName specifies the table name
func (ProductAttributeInt) TableName() string {
	return "catalog_product_entity_int"
}

/* Usage Examples:

1. Create:
   ```go
   attrInt := &ProductAttributeInt{
       AttributeID: 1,
       StoreID: 1,
       EntityID: 1,
       Value: 100,
   }
   db.Create(attrInt)
   ```

2. Read:
   ```go
   var attrInt ProductAttributeInt
   db.First(&attrInt, valueID)
   ```

3. Update:
   ```go
   db.Model(&attrInt).Update("Value", 200)
   ```

4. Delete:
   ```go
   db.Delete(&attrInt)
   ```
*/
