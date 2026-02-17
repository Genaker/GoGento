package product

import (
	"time"
	//entity "magento.GO/model/entity"
)

type StockItem struct {
	ItemID                  uint       `gorm:"column:item_id;primaryKey;autoIncrement"`
	ProductID               uint       `gorm:"column:product_id;not null"`
	StockID                 uint16     `gorm:"column:stock_id;not null"`
	Qty                     float64    `gorm:"column:qty"`
	MinQty                  float64    `gorm:"column:min_qty;not null;default:0.0000"`
	UseConfigMinQty         uint16     `gorm:"column:use_config_min_qty;not null;default:1"`
	IsQtyDecimal            uint16     `gorm:"column:is_qty_decimal;not null;default:0"`
	Backorders              uint16     `gorm:"column:backorders;not null;default:0"`
	UseConfigBackorders     uint16     `gorm:"column:use_config_backorders;not null;default:1"`
	MinSaleQty              float64    `gorm:"column:min_sale_qty;not null;default:1.0000"`
	UseConfigMinSaleQty     uint16     `gorm:"column:use_config_min_sale_qty;not null;default:1"`
	MaxSaleQty              float64    `gorm:"column:max_sale_qty;not null;default:0.0000"`
	UseConfigMaxSaleQty     uint16     `gorm:"column:use_config_max_sale_qty;not null;default:1"`
	IsInStock               uint16     `gorm:"column:is_in_stock;not null;default:0"`
	LowStockDate            *time.Time `gorm:"column:low_stock_date"`
	NotifyStockQty          *float64   `gorm:"column:notify_stock_qty"`
	UseConfigNotifyStockQty uint16     `gorm:"column:use_config_notify_stock_qty;not null;default:1"`
	ManageStock             uint16     `gorm:"column:manage_stock;not null;default:0"`
	UseConfigManageStock    uint16     `gorm:"column:use_config_manage_stock;not null;default:1"`
	StockStatusChangedAuto  uint16     `gorm:"column:stock_status_changed_auto;not null;default:0"`
	UseConfigQtyIncrements  uint16     `gorm:"column:use_config_qty_increments;not null;default:1"`
	QtyIncrements           float64    `gorm:"column:qty_increments;not null;default:0.0000"`
	UseConfigEnableQtyInc   uint16     `gorm:"column:use_config_enable_qty_inc;not null;default:1"`
	EnableQtyIncrements     uint16     `gorm:"column:enable_qty_increments;not null;default:0"`
	IsDecimalDivided        uint16     `gorm:"column:is_decimal_divided;not null;default:0"`
	WebsiteID               uint16     `gorm:"column:website_id;not null;default:0"`

	// Relationships
	//Product has a field of type StockItem
	//StockItem has a field of type Product
	//This creates an infinite loop in the type definitions, which Go does not allow.
	Product *Product `gorm:"foreignKey:ProductID;references:EntityID"`
	//Stock     entity.Stock           `gorm:"foreignKey:StockID;references:StockID"`

}

// TableName specifies the table name
func (StockItem) TableName() string {
	return "cataloginventory_stock_item"
}

/* Usage Examples:

1. Create:
   ```go
   stockItem := &StockItem{
       ProductID: 123,
       StockID: 1,
       Qty: 10.0,
       IsInStock: 1,
   }
   db.Create(stockItem)
   ```

2. Read with relationships:
   ```go
   var stockItem StockItem
   db.Preload("Product").Preload("Stock").First(&stockItem, id)
   ```

3. Update:
   ```go
   db.Model(&stockItem).Updates(map[string]interface{}{
       "qty": 5.0,
   })
   ```

4. Delete:
   ```go
   db.Delete(&stockItem)
   ```
*/
