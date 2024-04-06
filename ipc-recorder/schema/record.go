package schema

import "time"

type Record struct {
	ID        string    `gorm:"size:20;primaryKey" json:"id"`
	Device    string    `gorm:"size:20;index" json:"device"`
	Stream    string    `gorm:"size:20;index" json:"stream"`
	FileURL   string    `gorm:"size:255" json:"file_url"`
	FileSize  int64     `json:"file_size"`
	StartAt   time.Time `gorm:"index" json:"start_at"`
	EndAt     time.Time `gorm:"index" json:"end_at"`
	Interval  float64   `json:"interval"`
	Extra     JSONMap   `gorm:"type:jsonb;" json:"extra"` // Extra information (JSON)
	CreatedAt time.Time `gorm:"index;" json:"created_at"` // Created time
}
