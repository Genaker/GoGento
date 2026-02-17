package product

type ProductIndexPrice struct {
	EntityID        uint    `gorm:"column:entity_id;primaryKey"`
	CustomerGroupID uint    `gorm:"column:customer_group_id;primaryKey"`
	WebsiteID       uint16  `gorm:"column:website_id;primaryKey"`
	TaxClassID      uint16  `gorm:"column:tax_class_id;default:0"`
	Price           float64 `gorm:"column:price"`
	FinalPrice      float64 `gorm:"column:final_price"`
	MinPrice        float64 `gorm:"column:min_price"`
	MaxPrice        float64 `gorm:"column:max_price"`
	TierPrice       float64 `gorm:"column:tier_price"`
}

// TableName specifies the table name
func (ProductIndexPrice) TableName() string {
	return "catalog_product_index_price"
}

/* Usage Examples:

1. Create:
   ```go
   price := &ProductIndexPrice{
       EntityID: 123,
       CustomerGroupID: 1,
       WebsiteID: 1,
       Price: 99.99,
       FinalPrice: 89.99,
   }
   db.Create(price)
   ```

2. Read:
   ```go
   var price ProductIndexPrice
   db.First(&price, "entity_id = ? AND customer_group_id = ? AND website_id = ?", 123, 1, 1)
   ```

3. Update:
   ```go
   db.Model(&price).Update("final_price", 79.99)
   ```

4. Delete:
   ```go
   db.Delete(&price)
   ```
*/
