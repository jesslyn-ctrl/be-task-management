package model

type User struct {
	Base
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"-"` // Excluded from JSON for security
	Teams    []*Team `json:"teams" gorm:"many2many:user_teams"`
}
