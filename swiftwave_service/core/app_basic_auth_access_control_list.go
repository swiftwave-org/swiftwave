package core

import (
	"context"
	"errors"
	"github.com/oklog/ulid"
	"gorm.io/gorm"
	"math/rand"
	"strings"
	"time"
)

func FindAllAppBasicAuthAccessControlLists(_ context.Context, db *gorm.DB) ([]*AppBasicAuthAccessControlList, error) {
	var l []*AppBasicAuthAccessControlList
	if err := db.Find(&l).Error; err != nil {
		return nil, err
	}
	return l, nil
}

func (l *AppBasicAuthAccessControlList) FindByID(_ context.Context, db *gorm.DB, id uint) error {
	return db.First(l, id).Error
}

func (l *AppBasicAuthAccessControlList) Create(_ context.Context, db *gorm.DB) error {
	l.Name = strings.TrimSpace(l.Name)
	if strings.Compare(l.Name, "") == 0 {
		return errors.New("name cannot be empty")
	}
	l.GeneratedName = ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
	return db.Create(l).Error
}

func (l *AppBasicAuthAccessControlList) Delete(_ context.Context, db *gorm.DB) error {
	// TODO check if any app has been protected by this userlist
	return db.Delete(l).Error
}
