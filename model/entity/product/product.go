package product

import (
	"time"

	"magento.GO/model/entity/category"
)

// IsEnterprise is set at runtime by DetectEdition()
var IsEnterprise bool

// Product represents catalog_product_entity table.
// JSON tags use omitempty to exclude zero-value fields from serialization,
// keeping API responses compact (e.g., RowID=0 omitted for CE, EntityID=0 for EE if null).
type Product struct {
	EntityID        uint      `gorm:"column:entity_id;primaryKey;autoIncrement" json:"entity_id,omitempty"`
	RowID           uint      `gorm:"-" json:"row_id,omitempty"`
	AttributeSetID  uint16    `gorm:"column:attribute_set_id;type:smallint unsigned;not null;default:0" json:"attribute_set_id,omitempty"`
	TypeID          string    `gorm:"column:type_id;type:varchar(32);not null;default:simple" json:"type_id,omitempty"`
	SKU             string    `gorm:"column:sku;type:varchar(64);not null" json:"sku,omitempty"`
	HasOptions      uint16    `gorm:"column:has_options;type:smallint;not null;default:0" json:"has_options,omitempty"`
	RequiredOptions uint16    `gorm:"column:required_options;type:smallint unsigned;not null;default:0" json:"required_options,omitempty"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at,omitempty"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at,omitempty"`
	CreatedIn       uint64    `gorm:"-" json:"created_in,omitempty"`
	UpdatedIn       uint64    `gorm:"-" json:"updated_in,omitempty"`
	Categories         []category.Category   `gorm:"many2many:catalog_category_product;joinForeignKey:ProductID;joinReferences:CategoryID" json:"categories,omitempty"`
	Varchars           []ProductVarchar      `gorm:"foreignKey:EntityID;references:EntityID" json:"varchars,omitempty"`
	Ints               []ProductInt          `gorm:"foreignKey:EntityID;references:EntityID" json:"ints,omitempty"`
	Decimals           []ProductDecimal      `gorm:"foreignKey:EntityID;references:EntityID" json:"decimals,omitempty"`
	Texts              []ProductText         `gorm:"foreignKey:EntityID;references:EntityID" json:"texts,omitempty"`
	Datetimes          []ProductDatetime     `gorm:"foreignKey:EntityID;references:EntityID" json:"datetimes,omitempty"`
	MediaGallery       []ProductMediaGallery `gorm:"many2many:catalog_product_entity_media_gallery_value_to_entity;joinForeignKey:EntityID;joinReferences:ValueID" json:"media_gallery,omitempty"`
	StockItem          StockItem             `gorm:"foreignKey:EntityID;references:ProductID" json:"stock_item,omitempty"`
	ProductIndexPrices []ProductIndexPrice   `gorm:"foreignKey:EntityID;references:EntityID" json:"product_index_prices,omitempty"`
}

func (Product) TableName() string {
	return "catalog_product_entity"
}

// EAVLinkID returns the ID for EAV foreign keys (row_id for EE, entity_id for CE)
func (p *Product) EAVLinkID() uint {
	if IsEnterprise {
		return p.RowID
	}
	return p.EntityID
}
