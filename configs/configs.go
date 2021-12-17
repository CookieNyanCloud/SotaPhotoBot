package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
)

const (
	tgToken     = "TOKEN_A"
	DrivePeople = "DRIVEAPI_PEOPLE"
	DriveZag    = "DRIVEAPI_ZAG"
	filePath    = "users.json"
)

type Conf struct {
	TgToken  string
	DrivePpl string
	DriveZg  string
}

func InitConf() (*Conf, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return &Conf{}, err
	}
	return &Conf{
		TgToken:  os.Getenv(tgToken),
		DrivePpl: os.Getenv(DrivePeople),
		DriveZg:  os.Getenv(DriveZag),
	}, nil
}

func AddUser(users map[string]string, you, user string) error {
	users[you] = "найти"
	users[user] = "найти"
	jsonUsers, err := json.Marshal(users)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, jsonUsers, 0644)
	if err != nil {
		return err
	}
	return nil
}
