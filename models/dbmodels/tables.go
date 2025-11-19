package dbmodels

func Tables() []interface{} {
	return []interface{}{
		&User{},
		&Room{},
		&Player{},
		&Scores{},
	}
}
