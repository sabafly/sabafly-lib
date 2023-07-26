package logging

import (
	"bytes"
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
	o, err := os.OpenFile(filepath.Join(cfg.LogPath, cfg.LogName), os.O_RDWR|os.O_SYNC|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	var lines, seq int
	seq = 1
	fi, err := o.Stat()
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(o)
	if err != nil {
		return nil, err
	}
	lines = bytes.Count(buf, []byte("\n"))
	l := &Logging{
		config:   cfg,
		time:     fi.ModTime(),
		seq:      seq,
		lines:    lines,
		file:     o,
		fileInfo: fi,
	}
	if time.Now().Add(-3*time.Hour).After(l.time) || l.lines > 512 {
		if err := l.write(); err != nil {
			return nil, fmt.Errorf("error on write: %w", err)
		}
	}

	return l, nil
}

func (l *Logging) Close() error {
	return l.file.Close()
}

type Logging struct {
	sync.Mutex
	config   Config
	time     time.Time
	seq      int
	lines    int
	file     *os.File
	lastDate int
	fileInfo os.FileInfo
}

func (l *Logging) Levels() []logrus.Level {
	return l.config.LogLevels
}

func (l *Logging) Fire(entry *logrus.Entry) error {
	return l.Log(entry.Level.String(), entry.Message, entry.Time)
}

func (l *Logging) Log(lvl, message string, t time.Time) error {
	if (t.Add(-12*time.Hour).After(l.time) && l.lines > 0) || (t.Add(-3*time.Hour).After(l.time) && l.lines > 512) {
		if err := l.write(); err != nil {
			return err
		}
	}
	l.Lock()
	defer l.Unlock()
	if _, err := l.file.WriteString(fmt.Sprintf("[%s] [%s]: %s\n", t.Format(time.TimeOnly), lvl, message)); err != nil {
		return err
	}
	l.time = t
	l.lines++
	return nil
}

func (l *Logging) write() error {
	l.Lock()
	defer l.Unlock()
	if l.lastDate != time.Now().Day() {
		l.seq = 0
	}
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("error on sync: %w", err)
	}
	fi, err := l.file.Stat()
	if err != nil {
		return fmt.Errorf("error on file stat: %w", err)
	}
	path := fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), l.seq)
	if l.config.LogName != "latest.log" {
		path = strings.TrimSuffix(l.config.LogName, filepath.Ext(l.config.LogName)) + "-" + path
	}
	path = filepath.Join(l.config.LogPath, path)

	l.file.Close()

	os.Remove(filepath.Join(l.config.LogPath, l.config.LogName))

	o, err := os.OpenFile(filepath.Join(l.config.LogPath, l.config.LogName), os.O_RDWR|os.O_SYNC|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("error on os open: %w", err)
	}

	defer func() {
		l.file = o
	}()

	gz, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
	for err != nil {
		gz.Close()
		l.seq++
		path = fmt.Sprintf("%s-%d.gz", fi.ModTime().Format(time.DateOnly), l.seq)
		if l.config.LogName != "latest.log" {
			path = strings.TrimSuffix(l.config.LogName, filepath.Ext(l.config.LogName)) + "-" + path
		}
		path = filepath.Join(l.config.LogPath, path)
		gz, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
	}
	defer gz.Close()
	gw := gzip.NewWriter(gz)
	defer gw.Close()
	gw.Header = gzip.Header{
		Name:    l.config.LogName,
		ModTime: fi.ModTime(),
	}
	if _, err := io.Copy(gw, o); err != nil {
		return fmt.Errorf("error on io copy: %w", err)
	}

	if err := o.Truncate(0); err != nil {
		return fmt.Errorf("error on truncate: %w", err)
	}
	_, _ = o.Seek(0, io.SeekStart)

	l.lines = 0
	l.lastDate = time.Now().Day()
	return nil
}
