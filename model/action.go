package model

import "time"

type Action struct {
	ID        int64
	Guild     int64 `gorm:"not null"`
	CreatedBy int64
	CreatedAt time.Time

	LogChannel *int64
	RoleGrant  *int64
	RoleRevoke *int64
	Name       *string
	NameReset  *bool
	Kick       *bool
}
