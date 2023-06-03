package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func New(cfg Config) (*Logging, error) {
	_ = os.Mkdir(filepath.Clean(cfg.FilePath), os.ModeDir)
	o, err := os.Open(filepath.Join(cfg.FilePath, "latest.log"))
	if err == nil {
		fi, err := o.Stat()
		if err != nil {
			return nil, err
		}
		seq := 1
		_, err = os.Open(filepath.Join(cfg.FilePath, fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), seq)))
		for !os.IsNotExist(err) {
			seq++
			_, err = os.Open(filepath.Join(cfg.FilePath, fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), seq)))
		}
		tg, err := os.Create(filepath.Join(cfg.FilePath, fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), seq)))
		if err != nil {
			return nil, err
		}
		defer tg.Close()
		gw := gzip.NewWriter(tg)
		defer gw.Close()
		gw.Header = gzip.Header{
			Name:    "latest.log",
			ModTime: fi.ModTime(),
		}
		if _, err := io.Copy(gw, o); err != nil {
			return nil, err
		}
		o.Close()
	}
	f, err := os.Create(filepath.Join(cfg.FilePath, "latest.log"))
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &Logging{
		config:          cfg,
		fileCreatedTime: time.Now(),
		file:            f,
		fileInfo:        fi,
	}, nil
}

func (l *Logging) Close() error {
	return l.file.Close()
}

type Logging struct {
	sync.Mutex
	config          Config
	fileCreatedTime time.Time
	seq             int
	file            *os.File
	fileInfo        os.FileInfo
}

func (l *Logging) Levels() []logrus.Level {
	return l.config.LogLevels
}

func (l *Logging) Fire(entry *logrus.Entry) error {
	l.Lock()
	defer l.Unlock()
	if l.fileCreatedTime.Before(time.Now().Add(time.Hour * -3)) {
		if l.fileCreatedTime.Day() != entry.Time.Day() {
			l.seq = 0
		}
		l.seq++
		_, err := os.Open(filepath.Join(l.config.FilePath, fmt.Sprintf("%s-%d.gz", time.Now().Format(time.DateOnly), l.seq)))
		for !os.IsNotExist(err) {
			l.seq++
			_, err = os.Open(filepath.Join(l.config.FilePath, fmt.Sprintf("%s-%d.gz", time.Now().Format(time.DateOnly), l.seq)))
		}
		tg, err := os.Create(filepath.Join(l.config.FilePath, fmt.Sprintf("%s-%d.gz", time.Now().Format(time.DateOnly), l.seq)))
		if err != nil {
			return nil
		}
		defer tg.Close()
		gw := gzip.NewWriter(tg)
		defer gw.Close()
		gw.Header = gzip.Header{
			Name:    "latest.log",
			ModTime: l.fileInfo.ModTime(),
		}
		if _, err := io.Copy(gw, l.file); err != nil {
			return err
		}
		l.Close()
		_ = os.Remove(filepath.Join(l.config.FilePath, "latest.log"))
		f, err := os.Create(filepath.Join(l.config.FilePath, "latest.log"))
		if err != nil {
			return err
		}
		l.file = f
		fi, err := f.Stat()
		if err != nil {
			return err
		}
		l.fileInfo = fi
		l.fileCreatedTime = fi.ModTime()
	}
	_, err := l.file.WriteString(fmt.Sprintf("[%s] [%s]: %s\n", entry.Time.Format(time.TimeOnly), entry.Level.String(), entry.Message))
	if err != nil {
		return err
	}
	return nil
}
