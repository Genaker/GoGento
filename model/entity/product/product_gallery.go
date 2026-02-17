package product

import (
	"gorm.io/gorm"
	entity "magento.GO/model/entity"
	"time"
)

type ProductGallery struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;not null"`
	StoreID     uint16 `gorm:"column:store_id;not null"`
	EntityID    uint   `gorm:"column:entity_id;not null"`
	Position    int    `gorm:"column:position;not null;default:0"`
	Value       string `gorm:"column:value;type:varchar(255)"`

	// Relationships
	Attribute entity.EavAttribute `gorm:"foreignKey:AttributeID;references:AttributeID"`
	Product   Product             `gorm:"foreignKey:EntityID;references:EntityID"`
	//Store       entity.Store             `gorm:"foreignKey:StoreID;references:StoreID"`

	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name
func (ProductGallery) TableName() string {
	return "catalog_product_entity_gallery"
}

/* Usage Examples:

1. Create:
   ```go
   gallery := &ProductGallery{
       AttributeID: 73,
       StoreID: 1,
       EntityID: 123,
       Position: 0,
       Value: "/m/y/my-image.jpg",
   }
   db.Create(gallery)
   ```

2. Read with relationships:
   ```go
   var gallery ProductGallery
   db.Preload("Attribute").Preload("Product").Preload("Store").First(&gallery, id)
   ```

3. Update:
   ```go
   db.Model(&gallery).Updates(map[string]interface{}{
       "position": 1,
   })
   ```

4. Delete:
   ```go
   db.Delete(&gallery)
   ```
*/
