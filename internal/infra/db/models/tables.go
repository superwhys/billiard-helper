package models

func Tables() []interface{} {
	return []interface{}{
		&User{},
		&Room{},
		&Player{},
		&Score{},
	}
}
