package storage

import (
	"github.com/asdine/storm"
)

var DB *storm.DB

func Init() {
	db, err := storm.Open("contracts.db")
	if err != nil {
		panic(err)
	}

	DB = db
}
