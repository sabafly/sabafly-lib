package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func New(cfg Config) (*Logging, error) {
	_ = os.Mkdir(filepath.Clean(cfg.LogPath), 0755)
	if cfg.LogName == "" {
		cfg.LogName = "latest.log"
	}
	log_file, err := os.OpenFile(filepath.Join(cfg.LogPath, cfg.LogName), os.O_RDWR|os.O_SYNC|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(log_file)
	if err != nil {
		return nil, err
	}
	_, _ = log_file.Seek(0, 2)

	l := &Logging{
		config: cfg,
		file:   log_file,
		lines:  strings.Count(string(buf), "\r\n"),
	}

	go l.moveScheduler()

	return l, nil
}

func (l *Logging) moveScheduler() {
	tick := time.NewTicker(time.Hour)
	for {
		<-tick.C
		_ = l.move()
	}
}

var time_format string = "2006_01_02_15_04_05"

func (l *Logging) move() error {
	if l.lines < 1024 {
		return nil
	}
	l.Lock()
	defer l.Unlock()
	_, _ = l.file.Seek(0, 0)
	buf, err := io.ReadAll(l.file)
	if err != nil {
		return err
	}
	copy_file, err := os.OpenFile(filepath.Join(l.config.LogPath, fmt.Sprintf("%s%s.gz", l.config.Prefix, time.Now().Format(time_format))), os.O_RDWR|os.O_SYNC|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer copy_file.Close()
	gz := gzip.NewWriter(copy_file)
	defer gz.Close()
	if _, err := gz.Write(buf); err != nil {
		return err
	}

	if err := l.file.Truncate(0); err != nil {
		return err
	}
	_, _ = l.file.Seek(0, 0)

	l.lines = 0

	return nil
}

func (l *Logging) Close() error {
	return l.file.Close()
}

type Logging struct {
	sync.Mutex
	config Config
	lines  int
	file   *os.File
}

func (l *Logging) Levels() []logrus.Level {
	return l.config.LogLevels
}

func (l *Logging) Fire(entry *logrus.Entry) error {
	return l.Log(entry.Level.String(), entry.Message, entry.Time)
}

func (l *Logging) Log(lvl, message string, t time.Time) error {
	if l.lines > 2048 {
		_ = l.move()
	}
	l.Lock()
	defer l.Unlock()
	if _, err := l.file.WriteString(fmt.Sprintf("[%s] [%s]: %s\r\n", t.Format(time.DateTime), lvl, message)); err != nil {
		return err
	}
	l.lines++
	return nil
}
