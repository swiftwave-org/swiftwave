package core

import (
	"errors"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

func FetchAppBasicAuthAccessControlUsers(_ context.Context, db *gorm.DB, appBasicAuthAccessControlListID uint) ([]*AppBasicAuthAccessControlUser, error) {
	var users []*AppBasicAuthAccessControlUser
	if err := db.Where("app_basic_auth_access_control_list_id = ?", appBasicAuthAccessControlListID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (u *AppBasicAuthAccessControlUser) Create(_ context.Context, db *gorm.DB) error {
	// check if user exists under same user-list
	if db.Where("username = ? AND app_basic_auth_access_control_list_id = ?", u.Username, u.AppBasicAuthAccessControlListID).First(&AppBasicAuthAccessControlUser{}).RowsAffected > 0 {
		return errors.New("user already exists")
	}
	return db.Create(u).Error
}

func (u *AppBasicAuthAccessControlUser) Update(_ context.Context, db *gorm.DB) error {
	return db.Save(u).Error
}

func (u *AppBasicAuthAccessControlUser) Delete(_ context.Context, db *gorm.DB) error {
	return db.Delete(u).Error
}
