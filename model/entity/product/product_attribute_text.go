package product

type ProductAttributeText struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16 `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint   `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       string `gorm:"column:value;type:mediumtext"`
}

// TableName specifies the table name
func (ProductAttributeText) TableName() string {
	return "catalog_product_entity_text"
}

/* Usage Examples:

1. Create:
   ```go
   attrText := &ProductAttributeText{
       AttributeID: 1,
       StoreID: 1,
       EntityID: 1,
       Value: "example text",
   }
   db.Create(attrText)
   ```

2. Read:
   ```go
   var attrText ProductAttributeText
   db.First(&attrText, valueID)
   ```

3. Update:
   ```go
   db.Model(&attrText).Update("Value", "new text")
   ```

4. Delete:
   ```go
   db.Delete(&attrText)
   ```
*/
