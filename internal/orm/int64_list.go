package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Int64List []int64

func ToInt64List(values []uint32) Int64List {
	if len(values) == 0 {
		return Int64List{}
	}
	list := make(Int64List, len(values))
	for i, value := range values {
		list[i] = int64(value)
	}
	return list
}

func ToUint32List(values Int64List) []uint32 {
	if len(values) == 0 {
		return []uint32{}
	}
	list := make([]uint32, len(values))
	for i, value := range values {
		list[i] = uint32(value)
	}
	return list
}

func (list Int64List) Value() (driver.Value, error) {
	payload, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return string(payload), nil
}

func (list *Int64List) Scan(value any) error {
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
		return fmt.Errorf("unsupported Int64List type: %T", value)
	}
}
