package main

import (
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/gen"
)

//go:generate go run gen.go
func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "../../internal/infra/db/query",
		WithUnitTest:  false,
		FieldNullable: true,
		Mode:          gen.WithQueryInterface,
	})

	g.ApplyBasic(models.Tables()...)
	g.Execute()
}
