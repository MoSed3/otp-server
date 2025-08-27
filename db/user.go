package db

import (
	"errors"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	PhoneNumber string `gorm:"unique;not null;index:idx_user_phone_number;varchar(20)"`
	FirstName   string `gorm:"varchar(100)"`
	LastName    string `gorm:"varchar(100)"`
}

func CreateUser(tx *gorm.DB, number string) (*User, error) {
	user := &User{PhoneNumber: number}
	return user, tx.Create(user).Error
}

func GetUserByPhoneNumber(tx *gorm.DB, phoneNumber string) (*User, error) {
	user := &User{PhoneNumber: phoneNumber}
	return user, tx.Where(user).Find(user).Error
}

func (u *User) Update(tx *gorm.DB, firstName, lastName string) error {
	u.FirstName = firstName
	u.LastName = lastName
	return tx.Save(u).Error
}

func GetUserByID(tx *gorm.DB, id uint) (*User, error) {
	user := &User{}
	return user, tx.First(user, id).Error
}

func GetOrCreateUserByPhoneNumber(tx *gorm.DB, phoneNumber string) (*User, error) {
	user, err := GetUserByPhoneNumber(tx, phoneNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || user.ID == 0 {
			return CreateUser(tx, phoneNumber)
		}
		return nil, err
	}
	return user, nil
}
