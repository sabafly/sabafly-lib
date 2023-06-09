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
	_ = os.Mkdir(filepath.Clean(cfg.LogPath), os.ModeDir)
	if cfg.LogName == "" {
		cfg.LogName = "latest.log"
	}
	o, err := os.OpenFile(filepath.Join(cfg.LogPath, cfg.LogName), os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if _, err := o.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}
	if err == nil {
		buf, err := io.ReadAll(o)
		if err != nil {
			return nil, err
		}
		if strings.Count(string(buf), "\n") > 512 {
			fi, err := o.Stat()
			if err != nil {
				return nil, err
			}
			seq := 1
			path := fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), seq)
			if cfg.LogName != "latest.log" {
				path = cfg.LogName + "-" + path
			}
			path = filepath.Join(cfg.LogPath, path)
			_, err = os.Open(path)
			for !os.IsNotExist(err) {
				seq++
				o, err = os.Open(path)
				if err == nil {
					o.Close()
				}
			}
			tg, err := os.Create(path)
			if err != nil {
				return nil, err
			}
			defer tg.Close()
			gw := gzip.NewWriter(tg)
			defer gw.Close()
			gw.Header = gzip.Header{
				Name:    cfg.LogName,
				ModTime: fi.ModTime(),
			}
			if _, err := io.Copy(gw, o); err != nil {
				return nil, err
			}
			o.Close()
			o, err = os.Create(filepath.Join(cfg.LogPath, cfg.LogName))
			if err != nil {
				return nil, err
			}
		}
	}
	fi, err := o.Stat()
	if err != nil {
		return nil, err
	}

	return &Logging{
		config:          cfg,
		fileCreatedTime: time.Now(),
		seq:             0,
		lines:           0,
		file:            o,
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
	lines           int
	file            *os.File
	fileInfo        os.FileInfo
}

func (l *Logging) Levels() []logrus.Level {
	return l.config.LogLevels
}

func (l *Logging) Fire(entry *logrus.Entry) error {
	return l.Log(entry.Level.String(), entry.Message, entry.Time)
}

func (l *Logging) Log(lvl, message string, t time.Time) error {
	l.Lock()
	defer l.Unlock()
	if l.fileCreatedTime.Before(time.Now().Add(time.Hour*-3)) && l.lines > 512 {
		if l.fileCreatedTime.Day() != t.Day() {
			l.seq = 0
		}
		l.seq++
		path := fmt.Sprintf("%s-%d.gz", time.Now().Format(time.DateOnly), l.seq)
		if l.config.LogName != "latest.log" {
			path = l.config.LogName + "-" + path
		}
		path = filepath.Join(l.config.LogPath, path)
		_, err := os.Open(path)
		for !os.IsNotExist(err) {
			l.seq++
			_, err = os.Open(path)
		}
		tg, err := os.Create(path)
		if err != nil {
			return nil
		}
		defer tg.Close()
		gw := gzip.NewWriter(tg)
		defer gw.Close()
		gw.Header = gzip.Header{
			Name:    l.config.LogName,
			ModTime: l.fileInfo.ModTime(),
		}
		if _, err := io.Copy(gw, l.file); err != nil {
			return err
		}
		l.Close()
		_ = os.Remove(filepath.Join(l.config.LogPath, l.config.LogName))
		f, err := os.Create(filepath.Join(l.config.LogPath, l.config.LogName))
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
		l.lines = 0
	}
	_, err := l.file.WriteString(fmt.Sprintf("[%s] [%s]: %s\n", t.Format(time.TimeOnly), lvl, message))
	if err != nil {
		return err
	}
	l.lines++
	return nil
}
