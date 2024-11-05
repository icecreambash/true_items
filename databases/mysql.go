package databases

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var (
	once = sync.Once{}
	DB   *gorm.DB
)

func GetCon(tenant string) *gorm.DB {

	dsn := "root:root@tcp(localhost:3500)/ained_" + tenant + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	DB = db

	return DB
}
