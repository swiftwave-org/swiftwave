package core

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

// This file contains the operations for the Application model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

// FindAllUsers : find all users
func FindAllUsers(ctx context.Context, db gorm.DB) ([]User, error) {
	var users []User
	err := db.Find(&users).Error
	return users, err
}

// FindUserByID : find user by id
func FindUserByID(ctx context.Context, db gorm.DB, id uint) (User, error) {
	var user User
	err := db.First(&user, id).Error
	return user, err
}

// FindUserByUsername : find user by username
func FindUserByUsername(ctx context.Context, db gorm.DB, username string) (User, error) {
	var user User
	err := db.Where("username = ?", username).First(&user).Error
	return user, err
}

// CreateUser : create user
func CreateUser(ctx context.Context, db gorm.DB, user User) (User, error) {
	err := db.Create(&user).Error
	return user, err
}

// DeleteUserByID : delete user by id
func DeleteUserByID(ctx context.Context, db gorm.DB, id uint) error {
	err := db.Delete(&User{}, id).Error
	// Don't return error if record not found -- assume it's already deleted
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// DeleteUserByUsername : delete user by username
func DeleteUserByUsername(ctx context.Context, db gorm.DB, username string) error {
	err := db.Where("username = ?", username).Delete(&User{}).Error
	// Don't return error if record not found -- assume it's already deleted
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
