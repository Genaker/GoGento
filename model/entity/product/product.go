package product

import (
	"magento.GO/model/entity/category"
	"time"
)

type Product struct {
	EntityID           uint                  `gorm:"column:entity_id;primaryKey;autoIncrement"`
	AttributeSetID     uint16                `gorm:"column:attribute_set_id;type:smallint unsigned;not null;default:0"`
	TypeID             string                `gorm:"column:type_id;type:varchar(32);not null;default:simple"`
	SKU                string                `gorm:"column:sku;type:varchar(64);not null"`
	HasOptions         uint16                `gorm:"column:has_options;type:smallint;not null;default:0"`
	RequiredOptions    uint16                `gorm:"column:required_options;type:smallint unsigned;not null;default:0"`
	CreatedAt          time.Time             `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoCreateTime"`
	UpdatedAt          time.Time             `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
	Categories         []category.Category   `gorm:"many2many:catalog_category_product;joinForeignKey:ProductID;joinReferences:CategoryID"`
	Varchars           []ProductVarchar      `gorm:"foreignKey:EntityID;references:EntityID"`
	Ints               []ProductInt          `gorm:"foreignKey:EntityID;references:EntityID"`
	Decimals           []ProductDecimal      `gorm:"foreignKey:EntityID;references:EntityID"`
	Texts              []ProductText         `gorm:"foreignKey:EntityID;references:EntityID"`
	Datetimes          []ProductDatetime     `gorm:"foreignKey:EntityID;references:EntityID"`
	MediaGallery       []ProductMediaGallery `gorm:"many2many:catalog_product_entity_media_gallery_value_to_entity;joinForeignKey:EntityID;joinReferences:ValueID"`
	StockItem          StockItem             `gorm:"foreignKey:EntityID;references:ProductID"`
	ProductIndexPrices []ProductIndexPrice   `gorm:"foreignKey:EntityID;references:EntityID"`
}

// TableName specifies the table name
func (Product) TableName() string {
	return "catalog_product_entity"
}

/* Usage Examples:

1. Create:
   ```go
   product := &Product{
       AttributeSetID: 1,
       TypeID: "simple",
       SKU: "example_sku",
   }
   db.Create(product)
   ```

2. Read:
   ```go
   var product Product
   db.First(&product, productID)
   ```

3. Update:
   ```go
   db.Model(&product).Update("SKU", "new_sku")
   ```

4. Delete:
   ```go
   db.Delete(&product)
   ```
*/
