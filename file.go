package log4go

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

// --------------------------------------------------------------------------------
const (
	kDefaultLogFile = "temp_log.log"
	kLogFileExt     = ".log"
)

type FileWriterOption interface {
	Apply(*FileWriter)
}

type fwOptionFunc func(*FileWriter)

func (f fwOptionFunc) Apply(w *FileWriter) {
	f(w)
}

func WithMaxAge(sec int64) FileWriterOption {
	return fwOptionFunc(func(w *FileWriter) {
		if sec < 0 {
			return
		}
		w.maxAge = sec
	})
}

func WithMaxSize(mb int64) FileWriterOption {
	return fwOptionFunc(func(w *FileWriter) {
		if mb < 0 {
			return
		}
		w.maxSize = mb * 1024 * 1024
	})
}

func WithLogDir(dir string) FileWriterOption {
	return fwOptionFunc(func(w *FileWriter) {
		if dir == "" {
			return
		}
		w.dir = dir
	})
}

// --------------------------------------------------------------------------------
type FileWriter struct {
	level       int
	dir         string
	filename    string
	maxSize     int64
	maxAge      int64
	size        int64
	mu          sync.Mutex
	cmu         sync.Mutex
	file        *os.File
	enableColor bool
}

func NewFileWriter(level int, opts ...FileWriterOption) *FileWriter {
	var fw = &FileWriter{}
	fw.level = level
	fw.dir = "./logs"
	fw.maxSize = 10 * 1024 * 1024
	fw.maxAge = 0
	fw.enableColor = false
	for _, opt := range opts {
		opt.Apply(fw)
	}

	if err := os.MkdirAll(fw.dir, 0744); err != nil {
		return nil
	}
	fw.filename = path.Join(fw.dir, kDefaultLogFile)

	return fw
}

func (this *FileWriter) Level() int {
	return this.level
}

func (this *FileWriter) SetMaxSize(mb int) {
	this.maxSize = int64(mb) * 1024 * 1024
}

func (this *FileWriter) SetMaxAge(sec int64) {
	this.maxAge = sec
}

func (this *FileWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	pLen := int64(len(p))
	if this.file == nil {
		if err = this.openOrCreate(pLen); err != nil {
			return 0, err
		}
	}

	if this.size+pLen >= this.maxSize {
		if err := this.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = this.file.Write(p)
	this.size += int64(n)

	return n, err
}

func (this *FileWriter) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.close()
}

func (this *FileWriter) close() error {
	var err error
	if this.file != nil {
		err = this.file.Close()
	}
	this.file = nil
	this.size = 0
	return err
}

func (this *FileWriter) EnableColor() bool {
	return false
}

func (this *FileWriter) openOrCreate(pLen int64) error {
	this.cleanLogs()

	// 获取log文件信息
	info, err := os.Stat(this.filename)
	if os.IsNotExist(err) {
		// 如果log文件不存在，直接创建新的log文件
		return this.createFile()
	}
	if err != nil {
		return err
	}

	// 文件存在，但是其文件大小已超出设定的阈值
	if info.Size()+pLen >= this.maxSize {
		return this.rotate()
	}

	// 打开现有的文件
	file, err := os.OpenFile(this.filename, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		// 如果打开文件出错，则创建新的文件
		return this.createFile()
	}

	this.file = file
	this.size = info.Size()

	return nil
}

func (this *FileWriter) createFile() error {
	f, err := os.OpenFile(this.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	this.file = f
	this.size = 0
	return nil
}

func (this *FileWriter) renameFile() error {
	_, err := os.Stat(this.filename)
	if err == nil {
		var now = time.Now()
		var newName = path.Join(this.dir, fmt.Sprintf("log_%s_%d.log", now.Format("2006_01_02_15_04_05"), now.Nanosecond()))
		if err := os.Rename(this.filename, newName); err != nil {
			return err
		}
	}
	return err
}

func (this *FileWriter) rotate() error {
	if err := this.close(); err != nil {
		return err
	}

	if err := this.renameFile(); err != nil {
		return err
	}

	if err := this.createFile(); err != nil {
		return err
	}
	this.cleanLogs()
	return nil
}

func (this *FileWriter) cleanLogs() {
	if this.maxAge <= 0 {
		return
	}
	this.cmu.Lock()
	go func() {
		defer this.cmu.Unlock()
		var dir = filepath.Dir(this.dir)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) (rErr error) {
			defer func() {
				if r := recover(); r != nil {
				}
			}()

			if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-this.maxAge) {
				if filepath.Ext(info.Name()) == kLogFileExt && info.Name() != kDefaultLogFile {
					rErr = os.Remove(path)
				}
			}
			return rErr
		})
	}()
}
