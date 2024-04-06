package schema

import "time"

type Realtime struct {
	ID          string     `gorm:"size:20;primaryKey" json:"id"`
	Device      string     `gorm:"size:20;index" json:"device"`
	Stream      string     `gorm:"size:20;index" json:"stream"`
	LastStartAt *time.Time `json:"last_start_at"`
	SeqNo       int64      `json:"seq_no"`
	CreatedAt   time.Time  `gorm:"index;" json:"created_at"` // Created time
}
