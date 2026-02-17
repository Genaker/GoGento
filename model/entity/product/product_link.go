package product

type ProductLink struct {
	LinkID          uint   `gorm:"column:link_id;primaryKey;autoIncrement"`
	ProductID       uint   `gorm:"column:product_id;type:int unsigned;not null;default:0"`
	LinkedProductID uint   `gorm:"column:linked_product_id;type:int unsigned;not null;default:0"`
	LinkTypeID      uint16 `gorm:"column:link_type_id;type:smallint unsigned;not null;default:0"`
}

// TableName specifies the table name
func (ProductLink) TableName() string {
	return "catalog_product_link"
}

/* Usage Examples:

1. Create:
   ```go
   prodLink := &ProductLink{
       ProductID: 1,
       LinkedProductID: 2,
       LinkTypeID: 1,
   }
   db.Create(prodLink)
   ```

2. Read:
   ```go
   var prodLink ProductLink
   db.First(&prodLink, linkID)
   ```

3. Update:
   ```go
   db.Model(&prodLink).Update("LinkTypeID", 2)
   ```

4. Delete:
   ```go
   db.Delete(&prodLink)
   ```
*/
