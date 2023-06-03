package logging

import "github.com/sirupsen/logrus"

type Config struct {
	FilePath string `json:"file_path"`

	LogLevels []logrus.Level `json:"log_levels"`
}
