package model

import "time"

type User struct {
	ID        uint
	Username  string `gorm:"unique"`
	Email     string `gorm:"unique"`
	FullName  string
	Password  string
	Role      string
	CreatedAt time.Time

	 // Associations
    TeamMemberships []TeamMembership `gorm:"foreignKey:UserID" json:"-"`
    ContestRegistrations []ContestRegistration `gorm:"foreignKey:UserID" json:"-"`
    TrainingParticipations []TrainingParticipation `gorm:"foreignKey:UserID" json:"-"`
}
