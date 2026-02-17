package sales

import (
	"time"
)

type SalesOrderGrid struct {
	EntityID            uint       `gorm:"column:entity_id;primaryKey"`
	Status              string     `gorm:"column:status;type:varchar(32)"`
	StoreID             *uint      `gorm:"column:store_id"`
	StoreName           string     `gorm:"column:store_name;type:varchar(255)"`
	CustomerID          *uint      `gorm:"column:customer_id"`
	BaseGrandTotal      *float64   `gorm:"column:base_grand_total;type:decimal(20,4)"`
	BaseTotalPaid       *float64   `gorm:"column:base_total_paid;type:decimal(20,4)"`
	GrandTotal          *float64   `gorm:"column:grand_total;type:decimal(20,4)"`
	TotalPaid           *float64   `gorm:"column:total_paid;type:decimal(20,4)"`
	IncrementID         string     `gorm:"column:increment_id;type:varchar(50)"`
	BaseCurrencyCode    string     `gorm:"column:base_currency_code;type:varchar(3)"`
	OrderCurrencyCode   string     `gorm:"column:order_currency_code;type:varchar(255)"`
	ShippingName        string     `gorm:"column:shipping_name;type:varchar(255)"`
	BillingName         string     `gorm:"column:billing_name;type:varchar(255)"`
	CreatedAt           *time.Time `gorm:"column:created_at"`
	UpdatedAt           *time.Time `gorm:"column:updated_at"`
	BillingAddress      string     `gorm:"column:billing_address;type:varchar(255)"`
	ShippingAddress     string     `gorm:"column:shipping_address;type:varchar(255)"`
	ShippingInformation string     `gorm:"column:shipping_information;type:varchar(255)"`
	CustomerEmail       string     `gorm:"column:customer_email;type:varchar(255)"`
	CustomerGroup       string     `gorm:"column:customer_group;type:varchar(255)"`
	Subtotal            *float64   `gorm:"column:subtotal;type:decimal(20,4)"`
	ShippingAndHandling *float64   `gorm:"column:shipping_and_handling;type:decimal(20,4)"`
	CustomerName        string     `gorm:"column:customer_name;type:varchar(255)"`
	PaymentMethod       string     `gorm:"column:payment_method;type:varchar(255)"`
	TotalRefunded       *float64   `gorm:"column:total_refunded;type:decimal(20,4)"`
	PickupLocationCode  string     `gorm:"column:pickup_location_code;type:varchar(255)"`
	DisputeStatus       string     `gorm:"column:dispute_status;type:varchar(45)"`
	// Relationships (examples, actual models should be defined if needed)
	// Store              Store           `gorm:"foreignKey:StoreID"`
	// Customer           Customer        `gorm:"foreignKey:CustomerID"`
}

// TableName specifies the table name
func (SalesOrderGrid) TableName() string {
	return "sales_order_grid"
}

/* Usage Examples:

1. Create:
   ```go
   order := &SalesOrderGrid{
       Status: "processing",
       StoreID: ptrUint(1),
       CustomerID: ptrUint(123),
       GrandTotal: ptrFloat64(100.00),
       CreatedAt: ptrTime(time.Now()),
   }
   db.Create(order)
   ```

2. Read with relationships:
   ```go
   var order SalesOrderGrid
   db.First(&order, id)
   // db.Preload("Store").Preload("Customer").First(&order, id)
   ```

3. Update:
   ```go
   db.Model(&order).Updates(map[string]interface{}{
       "status": "complete",
   })
   ```

4. Delete:
   ```go
   db.Delete(&order)
   ```
*/
