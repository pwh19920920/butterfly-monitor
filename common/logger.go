package common

import (
	"context"
	"errors"
	"github.com/pwh19920920/butterfly/logger"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type loggerImpl struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
}

func NewGormLogger() *loggerImpl {
	return &loggerImpl{
		SkipErrRecordNotFound: true,
	}
}

func (l *loggerImpl) LogMode(gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *loggerImpl) Info(ctx context.Context, format string, args ...interface{}) {
	logger.InfoFormat(ctx, format, args...)
}

func (l *loggerImpl) Warn(ctx context.Context, format string, args ...interface{}) {
	logger.WarnFormat(ctx, format, args...)
}

func (l *loggerImpl) Error(ctx context.Context, format string, args ...interface{}) {
	logger.ErrorFormat(ctx, format, args...)
}

func (l *loggerImpl) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
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
