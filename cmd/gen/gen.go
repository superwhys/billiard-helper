package main

import (
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"gorm.io/gen"
)

//go:generate go run gen.go
func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "../../dal/db/query",
		WithUnitTest:  false,
		FieldNullable: true,
		Mode:          gen.WithQueryInterface,
	})

	g.ApplyBasic(dbmodels.Tables()...)
	g.Execute()
}
