package factory

import (
	"gorm.io/gorm"

	"learn-tbd/internal/pkg/database"
)

type Factory struct {
	Db *gorm.DB
}

func NewFactory() *Factory {

	f := &Factory{}

	f.SetupDb()

	return f
}

func (f *Factory) SetupDb() {
	db := database.Connection()
	f.Db = db
}

func (f *Factory) SetupRepository() {
	if f.Db == nil {
		panic("Failed setup repository, db is undefined")
	}
}
