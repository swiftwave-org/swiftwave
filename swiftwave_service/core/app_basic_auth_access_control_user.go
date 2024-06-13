package core

import (
	"errors"
	"gorm.io/gorm"
)

func (u *AppBasicAuthAccessControlUser) Create(db *gorm.DB) error {
	// check if user exists under same user-list
	if db.Where("username = ? AND app_basic_auth_access_control_list_id = ?", u.Username, u.AppBasicAuthAccessControlListID).First(&AppBasicAuthAccessControlUser{}).RowsAffected > 0 {
		return errors.New("user already exists")
	}
	return db.Create(u).Error
}

func (u *AppBasicAuthAccessControlUser) Update(db *gorm.DB) error {
	return db.Save(u).Error
}

func (u *AppBasicAuthAccessControlUser) Delete(db *gorm.DB) error {
	return db.Delete(u).Error
}
