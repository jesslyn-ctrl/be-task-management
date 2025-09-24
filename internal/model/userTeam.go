package model

type UserTeam struct {
	UserID string `json:"user_id" gorm:"primaryKey"`
	TeamID string `json:"team_id" gorm:"primaryKey"`
}
