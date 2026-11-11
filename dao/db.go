package dao

import (
	"database/sql"
	"os"

	"github.com/cntechpower/utils/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	persistDB *gorm.DB
	sqlDB     *sql.DB
	header    *log.Header
)

// Init opens the persistent SQLite database and runs migrations.
//
// SQLite has a single writer; MaxOpenConns(1) serializes access and, together
// with WAL + busy_timeout, avoids "database is locked" errors under load.
func Init(persistModels []interface{}) {
	header = log.NewHeader("db")
	initGorm(persistModels)
	header.Infof("GORM init finish")
}

func initGorm(persistModels []interface{}) {
	if len(persistModels) == 0 {
		return
	}
	var err error
	dsn := os.Getenv("ANYWHERE_DB_DSN")
	if dsn == "" {
		dsn = "proxy.db?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL"
	}
	persistDB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	sqlDB, err = persistDB.DB()
	if err != nil {
		panic("failed to get sql.DB: " + err.Error())
	}
	sqlDB.SetMaxOpenConns(1)
	if err := persistDB.AutoMigrate(persistModels...); err != nil {
		panic("failed to migrate database: " + err.Error())
	}
}

func Close() {
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
}

func PersistDB() *gorm.DB {
	if persistDB == nil {
		panic("PersistDB is not init")
	}
	return persistDB
}
