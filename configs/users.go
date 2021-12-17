package configs

import (
	"encoding/json"
	"os"
)

func GetUsers() (map[string]int, error) {
	var result map[string]int
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
