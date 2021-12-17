package configs

import (
	"encoding/json"
	"os"
)

func GetUsers() (map[string]string, error) {
	var result map[string]string
	bytes, err := os.ReadFile("configs/users.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
