package main

import (
	"github.com/rs/xid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"ipc-recorder/schema"
)

var db *gorm.DB

func initStorage() error {
	filename := "data/storage.db"
	dns := "file:" + filename + "?cache=private&mode=rwc"
	var err error
	db, err = gorm.Open(sqlite.Open(dns), &gorm.Config{})
	if err != nil {
		return err
	}

	// auto migration
	err = db.AutoMigrate(&schema.Record{})
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
