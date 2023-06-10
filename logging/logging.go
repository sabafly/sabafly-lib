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
	o, err := os.OpenFile(filepath.Join(cfg.LogPath, cfg.LogName), os.O_APPEND|os.O_RDWR|os.O_SYNC, os.ModeAppend)
	lines := 0
	seq := 1
	if err == nil {
		buf, err := io.ReadAll(o)
		if err != nil {
			return nil, err
		}
		lines = strings.Count(string(buf), "\n")
		if lines > 512 {
			fi, err := o.Stat()
			if err != nil {
				return nil, err
			}
			path := fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), seq)
			if cfg.LogName != "latest.log" {
				path = strings.TrimSuffix(cfg.LogName, filepath.Ext(cfg.LogName)) + "-" + path
			}
			path = filepath.Join(cfg.LogPath, path)
			fm, err := os.Open(path)
			if err != nil {
				fm.Close()
			}
			for !os.IsNotExist(err) || os.IsExist(err) {
				seq++
				path = fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), seq)
				if cfg.LogName != "latest.log" {
					path = strings.TrimSuffix(cfg.LogName, filepath.Ext(cfg.LogName)) + "-" + path
				}
				path = filepath.Join(cfg.LogPath, path)
				var f2 *os.File
				f2, err = os.Open(path)
				if err == nil {
					f2.Close()
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
			if _, err := gw.Write(buf); err != nil {
				return nil, fmt.Errorf("error on io copy: %w", err)
			}
			o.Close()
			lines = 0
			o, err = os.Create(filepath.Join(cfg.LogPath, cfg.LogName))
			if err != nil {
				return nil, fmt.Errorf("error on os crate: %w", err)
			}
		}
	} else {
		o, err = os.Create(filepath.Join(cfg.LogPath, cfg.LogName))
		if err != nil {
			return nil, fmt.Errorf("error on os crate: %w", err)
		}
	}
	_, _ = o.Seek(0, io.SeekEnd)
	fi, err := o.Stat()
	if err != nil {
		return nil, err
	}

	return &Logging{
		config:          cfg,
		fileCreatedTime: time.Now(),
		seq:             seq,
		lines:           lines,
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
	if l.lines > 512 {
		if l.fileCreatedTime.Day() != t.Day() {
			l.seq = 0
		}
		fi, err := l.file.Stat()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), l.seq)
		if l.config.LogName != "latest.log" {
			path = strings.TrimSuffix(l.config.LogName, filepath.Ext(l.config.LogName)) + "-" + path
		}
		path = filepath.Join(l.config.LogPath, path)
		fm, err := os.Open(path)
		if err != nil {
			fm.Close()
		}
		for !os.IsNotExist(err) || os.IsExist(err) {
			l.seq++
			path = fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), l.seq)
			if l.config.LogName != "latest.log" {
				path = strings.TrimSuffix(l.config.LogName, filepath.Ext(l.config.LogName)) + "-" + path
			}
			path = filepath.Join(l.config.LogPath, path)
			var f2 *os.File
			f2, err = os.Open(path)
			if err == nil {
				f2.Close()
			}
		}
		tg, err := os.Create(path)
		if err != nil {
			return err
		}
		defer tg.Close()
		gw := gzip.NewWriter(tg)
		defer gw.Close()
		gw.Header = gzip.Header{
			Name:    l.config.LogName,
			ModTime: fi.ModTime(),
		}
		if _, err := io.Copy(gw, l.file); err != nil {
			return fmt.Errorf("error on io copy: %w", err)
		}
		l.file.Close()
		l.lines = 0
		l.file, err = os.Create(filepath.Join(l.config.LogPath, l.config.LogName))
		if err != nil {
			return fmt.Errorf("error on os crate: %w", err)
		}
	}
	_, err := l.file.WriteString(fmt.Sprintf("[%s] [%s]: %s\n", t.Format(time.TimeOnly), lvl, message))
	if err != nil {
		return err
	}
	l.lines++
	return nil
}
