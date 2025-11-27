package models

func Tables() []any {
	return []any{
		&User{},
		&Match{},
		&Player{},
		&Score{},
	}
}
