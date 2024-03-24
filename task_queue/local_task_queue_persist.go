package task_queue

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/fatih/color"
	"gorm.io/gorm"
)

// EnqueuedTask holds the task details
// Add EnqueuedTask to gorm migration
type EnqueuedTask struct {
	ID        int `gorm:"primaryKey"`
	QueueName string
	Body      string
	Hash      string
}

func addTaskToDb(db *gorm.DB, queueName string, body string) error {
	h := sha256.New()
	_, err := h.Write([]byte(body))
	if err != nil {
		return err
	}
	hashSum := h.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	exists, err := existsTaskInDb(db, queueName, hashString)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	task := &EnqueuedTask{
		ID:        0,
		QueueName: queueName,
		Body:      body,
		Hash:      hashString,
	}
	result := db.Create(task)
	if result.Error != nil {
		color.Red(result.Error.Error())
		return result.Error
	}
	return nil
}

func existsTaskInDb(db *gorm.DB, queueName string, hash string) (bool, error) {
	var task EnqueuedTask
	result := db.Where("queue_name = ? AND hash = ?", queueName, hash).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func removeTaskFromDb(db *gorm.DB, queueName string, content string) error {
	h := sha256.New()
	_, err := h.Write([]byte(content))
	if err != nil {
		return err
	}
	hash := hex.EncodeToString(h.Sum(nil))
	result := db.Where("queue_name = ? AND hash = ?", queueName, hash).Delete(&EnqueuedTask{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func remoteTasksFromDb(db *gorm.DB, queueName string) error {
	result := db.Where("queue_name = ?", queueName).Delete(&EnqueuedTask{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func getTasksFromDb(db *gorm.DB, queueName string, removeTasks bool) (*[]EnqueuedTask, error) {
	var tasks []EnqueuedTask
	result := db.Where("queue_name = ?", queueName).Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	if removeTasks {
		_ = remoteTasksFromDb(db, queueName)
	}
	return &tasks, nil
}
