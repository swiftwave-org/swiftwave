package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"

	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
)

// FetchSystemLogRecords is the resolver for the fetchSystemLogRecords field.
func (r *queryResolver) FetchSystemLogRecords(ctx context.Context) ([]*model.FileInfo, error) {
	records, err := logger.FetchSystemLogRecords()
	if err != nil {
		return nil, err
	}
	var logFiles []*model.FileInfo
	for _, record := range records {
		logFiles = append(logFiles, &model.FileInfo{
			Name:    record.Name,
			ModTime: record.ModTime,
		})
	}
	if len(logFiles) > 0 {
		// sort the log files by mod time
		for i := 0; i < len(logFiles); i++ {
			for j := i + 1; j < len(logFiles); j++ {
				if logFiles[i].ModTime.Before(logFiles[j].ModTime) {
					logFiles[i], logFiles[j] = logFiles[j], logFiles[i]
				}
			}
		}
	}
	return logFiles, nil
}
