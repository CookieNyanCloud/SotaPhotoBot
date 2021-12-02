package configs

import (
	"os"
)

const (
	tgToken     = "TOKEN_A"
	DrivePeople = "DRIVEAPI_PEOPLE"
	DriveZag    = "DRIVEAPI_ZAG"
)

type Conf struct {
	TgToken  string
	DrivePpl string
	DriveZg  string
}

func InitConf() (*Conf, error) {
	return &Conf{
		TgToken:  os.Getenv(tgToken),
		DrivePpl: os.Getenv(DrivePeople),
		DriveZg:  os.Getenv(DriveZag),
	}, nil
}
