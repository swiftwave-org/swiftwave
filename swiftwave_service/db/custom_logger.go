package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	gormlogger "gorm.io/gorm/logger"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

// newCustomLogger initialize logger
func newCustomLogger(writer gormlogger.Writer, config gormlogger.Config) gormlogger.Interface {
	var (
		infoStr      = "%s: [info] "
		warnStr      = "%s: [warn] "
		errStr       = "%s: [error] "
		traceStr     = "%s: [trace] [%.3fms] [rows:%v] %s"
		traceWarnStr = "%s: [trace] %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s: [trace] %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		logger.DatabaseLogger.Printf("ignoring colorful logger")
	}

	return &customLogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type customLogger struct {
	gormlogger.Writer
	gormlogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *customLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	customLogger := *l
	customLogger.LogLevel = level
	return &customLogger
}

// Info print info
func (l *customLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.Printf(l.infoStr+msg, append([]interface{}{fetchCaller()}, data...)...)
	}
}

// Warn print warn messages
func (l *customLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{fetchCaller()}, data...)...)
	}
}

// Error print error messages
func (l *customLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.Printf(l.errStr+msg, append([]interface{}{fetchCaller()}, data...)...)
	}
}

// Trace print sql message
func (l *customLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceErrStr, fetchCaller(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceErrStr, fetchCaller(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, fetchCaller(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, fetchCaller(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormlogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, fetchCaller(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, fetchCaller(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter filter params
func (l *customLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}

func fetchCaller() string {
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		return ""
	}
	return filepath.Base(file) + ":" + strconv.Itoa(line)
}
