package gormkit

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type JSONObject map[string]interface{}

func (j *JSONObject) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal JSONB value:", value))
	}

	result := make(map[string]interface{})
	err := json.Unmarshal(bytes, &result)
	*j = JSONObject(result)
	return err
}

func (j JSONObject) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}

	return json.Marshal(j)
}

type JSONRaw json.RawMessage

func (j *JSONRaw) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSONRaw(result)
	return err
}

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}
