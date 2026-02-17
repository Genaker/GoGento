package entity

type EavAttribute struct {
	AttributeID    uint16  `gorm:"column:attribute_id;primaryKey;autoIncrement"`
	EntityTypeID   uint16  `gorm:"column:entity_type_id;type:smallint unsigned;not null;default:0"`
	AttributeCode  string  `gorm:"column:attribute_code;type:varchar(255);not null"`
	AttributeModel *string `gorm:"column:attribute_model;type:varchar(255)"`
	BackendModel   *string `gorm:"column:backend_model;type:varchar(255)"`
	BackendType    string  `gorm:"column:backend_type;type:varchar(8);not null;default:static"`
	BackendTable   *string `gorm:"column:backend_table;type:varchar(255)"`
	FrontendModel  *string `gorm:"column:frontend_model;type:varchar(255)"`
	FrontendInput  *string `gorm:"column:frontend_input;type:varchar(50)"`
	FrontendLabel  *string `gorm:"column:frontend_label;type:varchar(255)"`
	FrontendClass  *string `gorm:"column:frontend_class;type:varchar(255)"`
	SourceModel    *string `gorm:"column:source_model;type:varchar(255)"`
	IsRequired     uint16  `gorm:"column:is_required;type:smallint unsigned;not null;default:0"`
	IsUserDefined  uint16  `gorm:"column:is_user_defined;type:smallint unsigned;not null;default:0"`
	DefaultValue   *string `gorm:"column:default_value;type:text"`
	IsUnique       uint16  `gorm:"column:is_unique;type:smallint unsigned;not null;default:0"`
	Note           *string `gorm:"column:note;type:varchar(255)"`
}

func (EavAttribute) TableName() string {
	return "eav_attribute"
}

/* Usage Examples:

1. Create:
   attr := &EavAttribute{
       EntityTypeID: 4,
       AttributeCode: "name",
       BackendType: "varchar",
   }
   db.Create(attr)

2. Read:
   var attr EavAttribute
   db.First(&attr, 73)

3. Update:
   db.Model(&attr).Update("AttributeCode", "new_code")

4. Delete:
   db.Delete(&attr)
*/
