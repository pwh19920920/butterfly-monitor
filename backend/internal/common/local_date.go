package common

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type LocalDate struct {
	time.Time
}

func (t *LocalDate) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02"))
	return []byte(formatted), nil
}

func (t *LocalDate) UnmarshalJSON(data []byte) error {
	tt, _ := time.Parse(fmt.Sprintf("\"%s\"", "2006-01-02"), string(data))
	*t = LocalDate{tt}
	return nil
}

func (t *LocalDate) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *LocalDate) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = LocalDate{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
