package common

import (
	"database/sql/driver"
	"fmt"
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

	tt, err := time.ParseInLocation(fmt.Sprintf("\"%s\"", "2006-01-02 15:04:05"), string(data), time.Local)
	*t = LocalTime{tt}
	return err
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
