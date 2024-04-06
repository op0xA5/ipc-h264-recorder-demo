package schema

import (
	"database/sql/driver"
	"encoding/json"
)

type JSONMap map[string]interface{}

func (JSONMap) GormDataType() string {
	return "jsonb"
}

func (a *JSONMap) Scan(value interface{}) error {
	buf, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := make(JSONMap)
	if err := json.Unmarshal(buf, &result); err != nil {
		return err
	}
	*a = result
	return nil
}

func (a JSONMap) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a JSONMap) String() string {
	buf, _ := json.Marshal(a)
	return string(buf)
}

func (a JSONMap) CopyTo(dst interface{}) error {
	buf, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, dst)
}
