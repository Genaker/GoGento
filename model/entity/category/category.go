package category

import (
	"time"
)

type Category struct {
	EntityID       uint              `gorm:"column:entity_id;primaryKey;autoIncrement"`
	AttributeSetID uint16            `gorm:"column:attribute_set_id;type:smallint unsigned;not null;default:0"`
	ParentID       uint              `gorm:"column:parent_id;type:int unsigned;not null;default:0"`
	CreatedAt      time.Time         `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoCreateTime"`
	UpdatedAt      time.Time         `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
	Path           string            `gorm:"column:path;type:varchar(255);not null"`
	Position       int               `gorm:"column:position;not null"`
	Level          int               `gorm:"column:level;not null;default:0"`
	ChildrenCount  int               `gorm:"column:children_count;not null"`
	Products       []CategoryProduct `gorm:"foreignKey:CategoryID;references:EntityID"`
	Ints           []CategoryInt     `gorm:"foreignKey:EntityID;references:EntityID"`
	Varchars       []CategoryVarchar `gorm:"foreignKey:EntityID;references:EntityID"`
	Texts          []CategoryText    `gorm:"foreignKey:EntityID;references:EntityID"`
}

// TableName specifies the table name
func (Category) TableName() string {
	return "catalog_category_entity"
}

/* Usage Examples:

1. Create:
   ```go
   category := &Category{
       AttributeSetID: 1,
       Path: "1/2/3",
       Position: 1,
   }
   db.Create(category)
   ```

2. Read:
   ```go
   var category Category
   db.First(&category, categoryID)
   ```

3. Update:
   ```go
   db.Model(&category).Update("Path", "1/2/3/4")
   ```

4. Delete:
   ```go
   db.Delete(&category)
   ```
*/
