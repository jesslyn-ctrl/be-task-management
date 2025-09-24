package model

import "time"

type Task struct {
	Base
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"` // e.g., "To Do", "In Progress", "Done"
	DueDate     time.Time `json:"due_date"`

	AssignedTo   *string `json:"assigned_to"` // Foreign key to User
	AssignedUser *User   `json:"assigned_user" gorm:"foreignKey:AssignedTo"`

	TeamID string `json:"team_id"` // Foreign key to Team
	Team   *Team  `json:"team" gorm:"foreignKey:TeamID"`
}
