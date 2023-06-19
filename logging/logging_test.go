package logging_test

import (
	"testing"

	"github.com/sabafly/sabafly-lib/v2/logging"
	"github.com/sirupsen/logrus"
)

func TestLogging(t *testing.T) {
	lg := logrus.New()
	l, err := logging.New(logging.Config{
		LogPath:   "./__test__",
		LogName:   "test.log",
		LogLevels: logrus.AllLevels,
	})
	if err != nil {
		t.Error(err)
	}
	lg.AddHook(l)
	for i := 0; i < 512; i++ {
		lg.Info(i)
	}
}

func TestExistFile(t *testing.T) {
	if !t.Run("create log", TestLogging) {
		t.Fail()
	}
	lg := logrus.New()
	l, err := logging.New(logging.Config{
		LogPath:   "./__test__",
		LogName:   "test.log",
		LogLevels: logrus.AllLevels,
	})
	if err != nil {
		t.Error(err)
	}
	lg.AddHook(l)
	for i := 0; i < 520; i++ {
		lg.Info(i)
	}
}
