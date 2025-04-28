package model

import "time"

type Contest struct {
	ContestID   uint      `gorm:"primaryKey" json:"contest_id"`
	Name        string    `gorm:"type:varchar(100)" json:"name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsTeamBased bool      `gorm:"default:false" json:"is_team_based"`
	Organizer   string    `gorm:"type:varchar(100)" json:"organizer"`
	// Associations
	Registrations []ContestRegistration `gorm:"foreignKey:ContestID" json:"-"`
}

type ContestRegistration struct {
	RegistrationID     uint      `gorm:"primaryKey" json:"registration_id"`
	ContestID          uint      `gorm:"index" json:"contest_id"`
	IsUserRegistration bool      `gorm:"default:true" json:"is_user_registration"`
	UserID             *uint     `gorm:"index" json:"user_id,omitempty"`
	TeamID             *uint     `gorm:"index" json:"team_id,omitempty"`
	RegisteredAt       time.Time `json:"registered_at"`
	// Relations
	Contest *Contest `gorm:"foreignKey:ContestID;constraint:OnDelete:CASCADE" json:"-"`
	User    *User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Team    *Team    `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"-"`
}
