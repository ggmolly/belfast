package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringList []string

func (list StringList) Value() (driver.Value, error) {
	payload, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return string(payload), nil
}

func (list *StringList) Scan(value any) error {
	if value == nil {
		*list = nil
		return nil
	}
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), list)
	case []byte:
		return json.Unmarshal(v, list)
	default:
		return fmt.Errorf("unsupported StringList type: %T", value)
	}
}
