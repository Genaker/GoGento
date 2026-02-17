package product

import (
	"time"
	//"gorm.io/gorm"
	"gorm.io/datatypes"
)

type ProductJson struct {
	ID         uint           `gorm:"primaryKey;autoIncrement"`
	EntityID   uint           `gorm:"column:entity_id;uniqueIndex:unq_entity_store"` // Changed to uniqueIndex
	StoreID    uint           `gorm:"column:store_id;uniqueIndex:unq_entity_store;not null;default:0"`
	Attributes datatypes.JSON `gorm:"column:attribute_json;type:json not null"`

	// Timestamps
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationship
	Product Product `gorm:"foreignKey:EntityID;references:ID"`
}

func (ProductJson) TableName() string {
	return "product_json"
}

/* Usage Examples:

1. Create new product JSON entry:
productJson := &ProductJson{
	StoreID: 1,
	Attributes: datatypes.JSON(`{
		"sku": "24-MB01",
		"name": "Joust Duffle Bag",
		"description": "The perfect gym bag..."
	}`),
}
db.Create(productJson)

2. Query by SKU:
var results []ProductJson
db.Where("JSON_EXTRACT(attributes, '$.sku') = ?", "24-MB01").Find(&results)

3. Update attributes:
db.Model(&productJson).Update("Attributes", datatypes.JSON(`{
	"name": "Updated Bag Name"
}`))

4. Relationship to main Product (if exists):
type Product struct {
	ID      uint         `gorm:"primaryKey"`
	// ... other fields
	JsonData []ProductJson `gorm:"foreignKey:EntityID"`
}
*/
