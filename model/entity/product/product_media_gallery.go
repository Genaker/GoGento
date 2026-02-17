package product

import (
	//"gorm.io/gorm"
	entity "magento.GO/model/entity"
)

type ProductMediaGallery struct {
	ValueID     uint   `gorm:"column:value_id;primaryKey;autoIncrement"`
	AttributeID uint16 `gorm:"column:attribute_id;not null"`
	Value       string `gorm:"column:value;type:varchar(255)"`
	MediaType   string `gorm:"column:media_type;type:varchar(32);not null;default:'image'"`
	Disabled    uint16 `gorm:"column:disabled;not null;default:0"`

	// Relationships
	Attribute entity.EavAttribute `gorm:"foreignKey:AttributeID;references:AttributeID"`
}

// TableName specifies the table name
func (ProductMediaGallery) TableName() string {
	return "catalog_product_entity_media_gallery"
}

/* Usage Examples:

1. Create:
   ```go
   media := &ProductMediaGallery{
       AttributeID: 73,
       Value: "/m/y/my-image.jpg",
       MediaType: "image",
       Disabled: 0,
   }
   db.Create(media)
   ```

2. Read with relationships:
   ```go
   var media ProductMediaGallery
   db.Preload("Attribute").First(&media, id)
   ```

3. Update:
   ```go
   db.Model(&media).Updates(map[string]interface{}{
       "disabled": 1,
   })
   ```

4. Delete:
   ```go
   db.Delete(&media)
   ```
*/
