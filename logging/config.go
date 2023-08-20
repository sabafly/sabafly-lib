package logging

import "github.com/sirupsen/logrus"

type Config struct {
	LogPath string `json:"log_path"`
	LogName string `json:"log_name"`
	Prefix  string `json:"prefix"`

	LogLevels []logrus.Level `json:"log_levels"`
}
