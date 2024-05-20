package worker

import (
	"fmt"
	"strings"
)

type dockerPushLog struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Progress   string `json:"progress"`
	LineNumber int    `json:"lineNumber"`
}

type DockerPushLogger struct {
	idToLineNoMap       map[string]int
	log                 map[int]*dockerPushLog // map of line number to dockerPushLog
	noOflines           int
	linesPrinted        int
	isPrintingFirstTime bool
}

func newDockerPushLogger() *DockerPushLogger {
	return &DockerPushLogger{
		idToLineNoMap:       make(map[string]int),
		log:                 make(map[int]*dockerPushLog),
		noOflines:           0,
		linesPrinted:        0,
		isPrintingFirstTime: true,
	}
}

func (d *DockerPushLogger) push(log map[string]interface{}) string {
	if log["id"] == nil {
		return ""
	}
	id := log["id"].(string)
	status := "unknown"
	if log["status"] != nil {
		status = log["status"].(string)
	}
	progress := ""
	if log["progress"] != nil {
		progress = log["progress"].(string)
		if strings.Compare(progress, "") == 0 {
			progress = "[==================================================>]"
		}
	}
	if strings.Compare(status, "Pushed") == 0 || strings.Compare(status, "Layer already exists") == 0 {
		progress = "[==================================================>]"
	} else if strings.Compare(status, "Preparing") == 0 || strings.Compare(status, "Waiting") == 0 {
		progress = "[                                                   ]"
	}
	// check if id is in idToLineNoMap
	lineNo, ok := d.idToLineNoMap[id]
	if !ok {
		// insert new id record
		d.noOflines++
		d.idToLineNoMap[id] = d.noOflines
		d.log[d.noOflines] = &dockerPushLog{
			ID:         id,
			Status:     "",
			Progress:   "",
			LineNumber: d.noOflines,
		}
		lineNo = d.noOflines
	}
	// check if anyhow log record empty or null
	logRecord, ok := d.log[lineNo]
	if !ok {
		return ""
	}
	// update record
	if !strings.Contains(status, "digest") && !strings.Contains(status, "Mounted") {
		logRecord.Progress = progress
		logRecord.Status = status
	}
	// generate content
	return d.content(false)
}

func (d *DockerPushLogger) content(forDB bool) string {
	content := ""
	for i := 1; i <= d.noOflines; i++ {
		content += fmt.Sprintf("%s %s %s\r\n", d.log[i].ID, d.log[i].Progress, d.log[i].Status)
	}
	content += "\r\n"
	if forDB {
		content = "\r\n" + content
	} else {
		if !d.isPrintingFirstTime {
			// clear screen and move cursor to the beginning of the line
			content += "\u009b2J\u009b1;1H\r\n" + content
		} else {
			// insert 30 blank lines (as on client side xterm has 30 rows) and clear screen and move cursor to the beginning of the line
			content += "\u009b2J\u009b1;1H\r\n" + content
			d.isPrintingFirstTime = false
		}
	}

	d.linesPrinted = d.noOflines
	return content
}

func convertToUnicode(escapeSequence string) string {
	var unicodeString strings.Builder

	for i := 0; i < len(escapeSequence); i++ {
		unicodeString.WriteString(fmt.Sprintf("\\u%04x", escapeSequence[i]))
	}

	return unicodeString.String()
}
