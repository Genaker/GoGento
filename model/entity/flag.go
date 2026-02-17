package entity

import (
	"time"
)

type Flag struct {
	FlagID     uint      `gorm:"column:flag_id;primaryKey;autoIncrement"`
	FlagCode   string    `gorm:"column:flag_code;type:varchar(255);not null"`
	State      uint16    `gorm:"column:state;type:smallint unsigned;not null;default:0"`
	FlagData   string    `gorm:"column:flag_data;type:mediumtext"`
	LastUpdate time.Time `gorm:"column:last_update;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// TableName specifies the table name
func (Flag) TableName() string {
	return "flag"
}

/* Usage Examples:

1. Create:
   ```go
   flag := &Flag{
       FlagCode: "example_code",
       State: 1,
   }
   db.Create(flag)
   ```

2. Read:
   ```go
   var flag Flag
   db.First(&flag, flagID)
   ```

3. Update:
   ```go
   db.Model(&flag).Update("State", 2)
   ```

4. Delete:
   ```go
   db.Delete(&flag)
   ```
*/
