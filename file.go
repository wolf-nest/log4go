package log4go

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type FileWriter struct {
	level     int
	dir       string
	filename  string
	maxSize   int64
	maxAge    int64
	size      int64
	mutex     sync.Mutex
	file      *os.File
	cleanChan chan bool
}

func NewFileWriter(level int, logDir string) *FileWriter {
	var fw = &FileWriter{}
	fw.level = level
	fw.dir = logDir
	fw.filename = path.Join(logDir, "temp_logs.log")
	fw.maxSize = 10 * 1024 * 1024

	if err := os.MkdirAll(fw.dir, 0744); err != nil {
		return nil
	}

	fw.cleanChan = make(chan bool, 1)
	go fw.runClean()

	return fw
}

func (this *FileWriter) SetMaxSize(mb int) {
	this.maxSize = int64(mb) * 1024 * 1024
}

func (this *FileWriter) SetMaxAge(sec int64) {
	this.maxAge = sec
}

func (this *FileWriter) WriteMessage(msg *LogMessage) {
	if msg == nil {
		return
	}
	if msg.level < this.level {
		return
	}
	var out = fmt.Sprintf("%s %s [%s:%d] %s", msg.header, msg.levelName, msg.file, msg.line, msg.message)
	this.Write([]byte(out))
}

func (this *FileWriter) Close() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.close()
}

func (this *FileWriter) close() error {
	if this.file == nil {
		return nil
	}
	err := this.file.Close()
	this.file = nil
	this.size = 0
	return err
}

func (this *FileWriter) Write(p []byte) (n int, err error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

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

func (this *FileWriter) openOrCreate(writeLen int64) error {
	this.doClean()

	// 验证log文件信息
	info, err := os.Stat(this.filename)
	if os.IsNotExist(err) {
		return this.createFile()
	}
	if err != nil {
		return err
	}

	if info.Size()+writeLen >= this.maxSize {
		return this.rotate()
	}

	// 打开现有的文件
	file, err := os.OpenFile(this.filename, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return this.createFile()
	}

	this.file = file
	this.size = info.Size()

	return nil
}

func (this *FileWriter) createFile() error {
	_, err := os.Stat(this.filename)
	if err == nil {
		var now = time.Now()
		var newName = path.Join(this.dir, fmt.Sprintf("%s_%.9d.log", now.Format("2006_01_02_15_04_05"), now.Nanosecond()))
		if err := os.Rename(this.filename, newName); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(this.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	this.file = f
	this.size = 0
	return nil
}

func (this *FileWriter) rotate() error {
	if err := this.close(); err != nil {
		return err
	}
	if err := this.createFile(); err != nil {
		return err
	}
	this.doClean()
	return nil
}

func (this *FileWriter) cleanLogs() {
	if this.maxAge <= 0 {
		return
	}

	var dir = filepath.Dir(this.dir)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (rErr error) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-this.maxAge) {
			if filepath.Ext(info.Name()) == "log" {
				rErr = os.Remove(path)
			}
		}
		return rErr
	})
}

func (this *FileWriter) runClean() {
	for _ = range this.cleanChan {
		this.cleanLogs()
	}
}

func (this *FileWriter) doClean() {
	select {
	case this.cleanChan <- true:
	default:
	}
}
