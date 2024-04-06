package main

import (
	"github.com/rs/xid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"ipc-player/schema"
	"sort"
	"time"
)

var db *gorm.DB

func initStorage() error {
	filename := "../ipc-recorder/data/storage.db"
	dns := "file:" + filename + "?cache=private&mode=rwc"
	var err error
	db, err = gorm.Open(sqlite.Open(dns), &gorm.Config{})
	if err != nil {
		return err
	}

	// auto migration
	err = db.AutoMigrate(&schema.Record{}, &schema.Realtime{})
	if err != nil {
		return err
	}

	return nil
}

func CreateRecord(record *schema.Record) error {
	if record.ID == "" {
		record.ID = xid.New().String()
	}
	record.CreatedAt = record.CreatedAt.UTC().Truncate(0)
	return db.Create(record).Error
}

func QueryRecordByTimeRange(device string, stream string, start, end time.Time) ([]*schema.Record, error) {
	var records []*schema.Record
	err := db.Model(&schema.Record{}).Where("device = ? and stream = ? and (start_at >= ? and end_at <= ?)", device, stream, start, end).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func QueryRecordLast(device string, stream string, n int) ([]*schema.Record, error) {
	var records []*schema.Record
	err := db.Model(&schema.Record{}).Where("device = ? and stream = ?", device, stream).Order("start_at desc").Limit(n).Find(&records).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].StartAt.Before(records[j].StartAt)
	})

	return records, nil
}

func QueryRecordLastStartAt(device string, stream string, startAt time.Time, n int) ([]*schema.Record, error) {
	var records []*schema.Record
	err := db.Model(&schema.Record{}).Where("device = ? and stream = ? and start_at > ?", device, stream, startAt).Order("start_at asc").Limit(n).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func CreateRealtime(realtime *schema.Realtime) error {
	if realtime.ID == "" {
		realtime.ID = xid.New().String()
	}
	realtime.CreatedAt = realtime.CreatedAt.UTC().Truncate(0)
	return db.Create(realtime).Error
}

func GetRealtime(id string) (*schema.Realtime, error) {
	realtime := &schema.Realtime{}
	err := db.Model(&schema.Realtime{}).Where("id = ?", id).First(realtime).Error
	if err != nil {
		return nil, err
	}
	return realtime, nil
}

func SaveRealtime(realtime *schema.Realtime) error {
	return db.Save(realtime).Error
}
