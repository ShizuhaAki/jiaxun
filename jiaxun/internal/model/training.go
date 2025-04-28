package model

import "time"

type TrainingPlan struct {
	TrainingPlanID uint      `gorm:"primaryKey" json:"training_plan_id"`
	Title          string    `gorm:"type:varchar(100)" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	// Associations
	Participations []TrainingParticipation `gorm:"foreignKey:TrainingPlanID" json:"-"`
}

type TrainingParticipation struct {
	ParticipationID uint      `gorm:"primaryKey" json:"participation_id"`
	TrainingPlanID  uint      `gorm:"index" json:"training_plan_id"`
	UserID          *uint     `gorm:"index" json:"user_id,omitempty"`
	TeamID          *uint     `gorm:"index" json:"team_id,omitempty"`
	JoinedAt        time.Time `json:"joined_at"`
	// Relations
	TrainingPlan *TrainingPlan `gorm:"foreignKey:TrainingPlanID;constraint:OnDelete:CASCADE" json:"-"`
	User         *User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Team         *Team         `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"-"`
}
