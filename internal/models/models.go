package models

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserStatus int

const (
	UserStatusActive UserStatus = iota + 1
	UserStatusDisabled
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusActive:
		return "Active"
	case UserStatusDisabled:
		return "Disabled"
	default:
		return "Unknown"
	}
}

func (s UserStatus) Int() int {
	return int(s)
}

type User struct {
	gorm.Model
	PhoneNumber string     `gorm:"unique;not null;index:idx_user_phone_number;varchar(20)"`
	FirstName   string     `gorm:"varchar(100)"`
	LastName    string     `gorm:"varchar(100)"`
	Status      UserStatus `gorm:"default:1"`
}

type UserOtp struct {
	gorm.Model
	Code   string       `gorm:"not null"`
	UserID uint         `gorm:"not null"`
	User   *User        `gorm:"foreignkey:UserID;not null"`
	UsedAt sql.NullTime `gorm:""`
}

func (o UserOtp) Waste(tx *gorm.DB) error {
	o.UsedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}

	return tx.Save(o).Error
}

type Setting struct {
	ID                uint   `gorm:"primaryKey"`
	SecretKey         string `gorm:"not null"`
	AccessTokenExpire uint   `gorm:"not null"` // Minutes
}

type AdminRole int

const (
	RoleSuperAdmin AdminRole = iota + 1
	RoleSudoAdmin
	RoleVisitorAdmin
)

func (role AdminRole) Int() int {
	return int(role)
}

func (role AdminRole) String() string {
	switch role {
	case RoleSuperAdmin:
		return "Super"
	case RoleSudoAdmin:
		return "Sudo"
	case RoleVisitorAdmin:
		return "Visitor"
	default:
		return "Unknown"
	}
}

func (role AdminRole) IsValid() bool {
	switch role {
	case RoleSuperAdmin, RoleSudoAdmin, RoleVisitorAdmin:
		return true
	default:
		return false
	}
}

func ParseAdminRole(roleStr string) (AdminRole, error) {
	switch roleStr {
	case "Super":
		return RoleSuperAdmin, nil
	case "Sudo":
		return RoleSudoAdmin, nil
	case "Visitor":
		return RoleVisitorAdmin, nil
	default:
		return 0, errors.New("invalid admin role")
	}
}

type Admin struct {
	gorm.Model
	Username        string    `gorm:"unique;index;not null;"`
	Role            AdminRole `gorm:"not null;default:3"`
	HashedPassword  string    `gorm:"not null"`
	PasswordResetAt sql.NullTime
}

type UserSearchParams struct {
	ID          *uint       `schema:"id"`
	PhoneNumber *string     `schema:"phone_number"`
	FirstName   *string     `schema:"first_name"`
	LastName    *string     `schema:"last_name"`
	Status      *UserStatus `schema:"status"`
	Limit       int         `schema:"limit"`
	Offset      int         `schema:"offset"`
	SortBy      string      `schema:"sort_by"`
	SortOrder   string      `schema:"sort_order"`
}

func (p *UserSearchParams) SetDefaults() {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 10 // Default limit
	}
	if p.Offset < 0 {
		p.Offset = 0 // Default offset
	}

	validSortBy := map[string]bool{
		"id":           true,
		"phone_number": true,
		"first_name":   true,
		"last_name":    true,
		"status":       true,
	}
	if !validSortBy[p.SortBy] {
		p.SortBy = "id" // Default sort by ID
	}

	validSortOrder := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if !validSortOrder[p.SortOrder] {
		p.SortOrder = "asc" // Default sort order ascending
	}
}

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusDisabled:
		return true
	default:
		return false
	}
}
