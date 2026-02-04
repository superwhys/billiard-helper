package models

func Tables() []any {
	return []any{
		&User{},
		&Match{},
		&MatchGame{},
		&Player{},
		&Score{},
	}
}
