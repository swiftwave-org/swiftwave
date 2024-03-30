package core

import "gorm.io/gorm"

// FetchServerLogByServerID is a function to fetch all server logs from the database by server id.
// This will not send the log content.
func FetchServerLogByServerID(db *gorm.DB, serverID uint) ([]ServerLog, error) {
	var serverLogs []ServerLog
	err := db.Select("id", "title", "server_id", "created_at", "updated_at").Where("server_id = ?", serverID).Order("created_at DESC").Find(&serverLogs).Error
	return serverLogs, err
}

// FetchServerLogContentByID is a function to fetch the content of a server log from the database by its id.
func FetchServerLogContentByID(db *gorm.DB, id uint) (string, error) {
	var serverLog ServerLog
	err := db.Select("content").First(&serverLog, id).Error
	return serverLog.Content, err
}

// CreateServerLog is a function to create a new server log in the database.
func CreateServerLog(db *gorm.DB, serverLog *ServerLog) error {
	return db.Create(serverLog).Error
}

// FetchServerLogByID is a function to fetch a single server log from the database by its id.
func FetchServerLogByID(db *gorm.DB, id uint) (*ServerLog, error) {
	var serverLog ServerLog
	err := db.First(&serverLog, id).Error
	return &serverLog, err
}

// Update is a function to update a server log in the database.
func (serverLog *ServerLog) Update(db *gorm.DB) error {
	return db.Save(serverLog).Error
}
