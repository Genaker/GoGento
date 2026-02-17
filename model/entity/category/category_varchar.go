package category

type CategoryVarchar struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16 `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint   `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       string `gorm:"column:value;type:varchar(255)"`

	// Relationship
	// Category Category `gorm:"foreignKey:EntityID;references:EntityID"`
}

func (CategoryVarchar) TableName() string {
	return "catalog_category_entity_varchar"
}

/* Usage Examples:

1. Create:
   ```go
   attr := &CategoryVarchar{
       AttributeID: 42,
       StoreID: 1,
       EntityID: 123,
       Value: "Sample Value",
   }
   db.Create(attr)
   ```

2. Read with relationship:
   ```go
   var attr CategoryVarchar
   db.Preload("Category").First(&attr, id)
   ```

3. Update:
   ```go
   db.Model(&attr).Update("value", "New Value")
   ```

4. Delete:
   ```go
   db.Delete(&attr)
   ```
*/
