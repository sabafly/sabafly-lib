package logging

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
)

func (l *Logging) Move() error {
	return l.move()
}

func GetTargetPath(t *testing.T, relativePathFromProjectRoot string) string {
	_, testSourceFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(testSourceFile)
	return filepath.Join(currentDir, relativePathFromProjectRoot)
}

func TestMove(t *testing.T) {
	time_format = "2006_01_02_15_04_05__"

	cfg := Config{
		LogPath: GetTargetPath(t, "./"),
		LogName: "__test__.log",
		Prefix:  "__",

		LogLevels: logrus.AllLevels,
	}
	l, err := New(cfg)
	if err != nil {
		panic(err)
	}
	logger := logrus.New()
	logger.AddHook(l)
	for i := 0; i < 3000; i++ {
		logger.Info(i)
	}
	if err := l.Move(); err != nil {
		panic(err)
	}
	for i := 3000; i > 0; i-- {
		logger.Info(i)
	}
}
