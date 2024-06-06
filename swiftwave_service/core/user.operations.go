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
	if user.Username == "" {
		return User{}, errors.New("username cannot be empty")
	}
	if user.PasswordHash == "" {
		return User{}, errors.New("password cannot be empty")
	}
	err := db.Create(&user).Error
	return user, err
}

// DeleteUser : delete user by id
func DeleteUser(ctx context.Context, db gorm.DB, id uint) error {
	err := db.Delete(&User{}, id).Error
	// Don't return error if record not found -- assume it's already deleted
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// ChangePassword : change user password
func ChangePassword(ctx context.Context, db gorm.DB, username string, oldPassword string, newPassword string) error {
	// Fetch user
	user, err := FindUserByUsername(ctx, db, username)
	if err != nil {
		return errors.New("user not found")
	}
	// Check old password
	isCorrect := user.CheckPassword(oldPassword)
	if !isCorrect {
		return errors.New("old password is incorrect")
	}
	// Set new password
	err = user.SetPassword(newPassword)
	if err != nil {
		return errors.New("failed to set new password")
	}
	// Update user
	err = db.Save(&user).Error
	return err
}

// DisableTotp : disable Totp for user
func DisableTotp(ctx context.Context, db gorm.DB, id uint) error {
	user, err := FindUserByID(ctx, db, id)
	if err != nil {
		return errors.New("user not found")
	}
	// disable TOTP
	user.TotpEnabled = false
	err = db.Save(&user).Error
	return err
}
