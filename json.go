package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func read_json_value(path, field string) (string, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	var parsed_data map[string]interface{}
	err = json.Unmarshal([]byte(data), &parsed_data)

	if err != nil {
		return "", err
	}

	if result, ok := parsed_data[field].(string); ok {
		return result, nil
	}

	return "", fmt.Errorf("%v is not a field", field)
}

func write_json_value(path, field, value string) error {

	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	var parsed_data map[string]interface{}
	err = json.Unmarshal([]byte(data), &parsed_data)

	if err != nil {
		return err
	}

	parsed_data[field] = value

	updated_data, err := json.MarshalIndent(parsed_data, "", "  ")

	if err != nil {
		return err
	}

	err = os.WriteFile(path, updated_data, 0644)

	if err != nil {
		return err
	}

	return nil
}


