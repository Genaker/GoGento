package category

type CategoryInt struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;type:smallint unsigned;not null;default:0"`
	StoreID     uint16 `gorm:"column:store_id;type:smallint unsigned;not null;default:0"`
	EntityID    uint   `gorm:"column:entity_id;type:int unsigned;not null;default:0"`
	Value       int    `gorm:"column:value"`

	// Relationship
	// Category Category `gorm:"foreignKey:EntityID;references:EntityID"`
}

func (CategoryInt) TableName() string {
	return "catalog_category_entity_int"
}

/* Usage Examples:

1. Create:
   ```go
   attr := &CategoryInt{
       AttributeID: 42,
       StoreID: 1,
       EntityID: 123,
       Value: 7,
   }
   db.Create(attr)
   ```

2. Read with relationship:
   ```go
   var attr CategoryInt
   db.Preload("Category").First(&attr, id)
   ```

3. Update:
   ```go
   db.Model(&attr).Update("value", 99)
   ```

4. Delete:
   ```go
   db.Delete(&attr)
   ```
*/
