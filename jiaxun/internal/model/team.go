package model

import (
	"time"
)

type Team struct {
    TeamID    uint      `gorm:"primaryKey" json:"team_id"`
    TeamName  string    `gorm:"type:varchar(100)" json:"team_name"`
    CreatedAt time.Time `json:"created_at"`
    // Associations
    TeamMemberships []TeamMembership `gorm:"foreignKey:TeamID" json:"-"`
    ContestRegistrations []ContestRegistration `gorm:"foreignKey:TeamID" json:"-"`
    TrainingParticipations []TrainingParticipation `gorm:"foreignKey:TeamID" json:"-"`
}

type TeamMembership struct {
    UserID    uint   `gorm:"primaryKey" json:"user_id"`
    TeamID    uint   `gorm:"primaryKey" json:"team_id"`
    Role      string `gorm:"type:enum('member','captain');default:'member'" json:"role"`
    JoinedAt  time.Time `json:"joined_at"`
    // Relations
    User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
    Team *Team `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"-"`
}


