package models

func Tables() []any {
	return []any{
		&User{},
		&BlockedIdentity{},
		&Match{},
		&MatchGame{},
		&Player{},
		&MatchEvent{},
		&Feedback{},
	}
}
