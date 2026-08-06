package common

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type LocalTime struct {
	time.Time
}

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	// 空值不进行解析
	if len(data) == 2 {
		*t = LocalTime{time.Time{}}
		return nil
	}

	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*t = LocalTime{time.Time{}}
		return nil
	}

	// 兼容多种时间格式：标准年月日时分秒、ISO/RFC3339、纯日期
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if tt, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			*t = LocalTime{tt}
			return nil
		}
	}
	return fmt.Errorf("can not parse %q to LocalTime", s)
}

func (t *LocalTime) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = LocalTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
