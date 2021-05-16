package models

import "gorm.io/gorm"

type Login string
type SessionID string
type Password string

type User struct {
	gorm.Model
	Login    Login `gorm:"uniqueIndex"`
	Password Password
	IsAdmin  bool
}

type Session struct {
	gorm.Model
	SessionID SessionID `gorm:"uniqueIndex"`
	UserID    uint
}
