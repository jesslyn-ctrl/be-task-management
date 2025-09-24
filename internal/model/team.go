package model

type Team struct {
	Base
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Users       []*User `json:"users" gorm:"many2many:user_teams"`
}
