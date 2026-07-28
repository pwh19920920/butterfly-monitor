package common

import (
	"context"
	"errors"
	"time"

	"github.com/pwh19920920/butterfly/pkg/logger"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type LoggerImpl struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
}

func NewGormLogger() *LoggerImpl {
	return &LoggerImpl{
		SkipErrRecordNotFound: true,
	}
}

func (l *LoggerImpl) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *LoggerImpl) Info(ctx context.Context, format string, args ...interface{}) {
	logger.InfoFormat(ctx, format, args...)
}

func (l *LoggerImpl) Warn(ctx context.Context, format string, args ...interface{}) {
	logger.WarnFormat(ctx, format, args...)
}

func (l *LoggerImpl) Error(ctx context.Context, format string, args ...interface{}) {
	logger.ErrorFormat(ctx, format, args...)
}

func (l *LoggerImpl) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := log.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[log.ErrorKey] = err
		logger.ErrorEntryFormat(ctx, log.WithFields(fields), "%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		logger.ErrorEntryFormat(ctx, log.WithFields(fields), "%s [%s]", sql, elapsed)
		return
	}
	logger.DebugEntryFormat(ctx, log.WithFields(fields), "%s [%s]", sql, elapsed)
}
