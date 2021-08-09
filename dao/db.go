package dao

import (
	"github.com/cntechpower/utils/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var persistDB *gorm.DB
var memDB *gorm.DB
var header *log.Header

func Init(persistModels []interface{}, tmpModels []interface{}) {
	header = log.NewHeader("db")
	initGorm(persistModels, tmpModels)
	header.Infof("GORM init finish")
}

func initGorm(persistModels []interface{}, tmpModels []interface{}) {
	var err error
	if len(persistModels) != 0 {
		persistDB, err = gorm.Open(sqlite.Open("proxy.db"), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		_ = persistDB.AutoMigrate(persistModels...)
	}

	if len(tmpModels) != 0 {
		memDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		_ = memDB.AutoMigrate(tmpModels...)
	}

}

func Close() {
	return
}

func PersistDB() *gorm.DB {
	if persistDB == nil {
		panic("PersistDB is not init")
	}
	return persistDB
}

func MemDB() *gorm.DB {
	if memDB == nil {
		panic("MemDB is not init")
	}
	return memDB
}
