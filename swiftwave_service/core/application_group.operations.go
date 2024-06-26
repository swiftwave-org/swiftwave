package core

import (
	"context"
	"errors"
	"github.com/dgryski/trifles/uuid"
	"gorm.io/gorm"
	"strings"
)

// This file contains the operations for the ApplicationGroup model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindAllApplicationGroups(_ context.Context, db gorm.DB) ([]*ApplicationGroup, error) {
	var groups []*ApplicationGroup
	err := db.Model(&ApplicationGroup{}).Scan(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, err
}

func FindApplicationsByApplicationGroupID(_ context.Context, db gorm.DB, groupId string) ([]*Application, error) {
	var applications []*Application
	err := db.Model(&Application{}).Where("application_group_id = ?", groupId).Scan(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (applicationGroup *ApplicationGroup) FindById(ctx context.Context, db gorm.DB, id string) error {
	return db.Where("id = ?", id).First(applicationGroup).Error
}

func (applicationGroup *ApplicationGroup) Create(_ context.Context, db gorm.DB) error {
	if strings.Compare(applicationGroup.ID, "") == 0 {
		applicationGroup.ID = uuid.UUIDv4()
	}
	if strings.Compare(applicationGroup.Name, "") == 0 {
		return errors.New("name cannot be blank")
	}
	return db.Create(applicationGroup).Error
}

func (applicationGroup *ApplicationGroup) Delete(_ context.Context, db gorm.DB) error {
	// check if no application is associated with this group
	applications, err := FindApplicationsByApplicationGroupID(context.Background(), db, applicationGroup.ID)
	if err != nil {
		return err
	}
	if len(applications) > 0 {
		return errors.New("application group has applications associated with it")
	}
	// delete group
	return db.Delete(applicationGroup).Error
}

func (applicationGroup *ApplicationGroup) IsAnyApplicationAssociatedWithGroup(ctx context.Context, db gorm.DB) (bool, error) {
	// verify from database
	var count int64
	tx := db.Model(&Application{}).Where("application_group_id = ?", applicationGroup.ID).Count(&count)
	if tx.Error != nil {
		return false, tx.Error
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
