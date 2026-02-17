package category

type CategoryProduct struct {
	EntityID   uint `gorm:"column:entity_id;primaryKey;autoIncrement"`
	CategoryID uint `gorm:"column:category_id;type:int unsigned;not null;default:0"`
	ProductID  uint `gorm:"column:product_id;type:int unsigned;not null;default:0"`
	Position   int  `gorm:"column:position;not null;default:0"`
}

// TableName specifies the table name
func (CategoryProduct) TableName() string {
	return "catalog_category_product"
}

/* Usage Examples:

1. Create:
   ```go
   catProd := &CategoryProduct{
       CategoryID: 1,
       ProductID: 1,
       Position: 1,
   }
   db.Create(catProd)
   ```

2. Read:
   ```go
   var catProd CategoryProduct
   db.First(&catProd, entityID)
   ```

3. Update:
   ```go
   db.Model(&catProd).Update("Position", 2)
   ```

4. Delete:
   ```go
   db.Delete(&catProd)
   ```
*/
