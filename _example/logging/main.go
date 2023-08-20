package main

import (
	"github.com/sabafly/sabafly-lib/v2/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := logging.Config{
		LogPath: "_example/logging",
		LogName: "latest.log",

		LogLevels: logrus.AllLevels,
	}
	l, err := logging.New(cfg)
	if err != nil {
		panic(err)
	}
	logger := logrus.New()
	logger.AddHook(l)
	for i := 0; i < 3000; i++ {
		logger.Info(i)
	}
}
