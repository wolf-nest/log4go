package log4go

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrFileClosed     = errors.New("file already closed")
	ErrFileBufferFull = errors.New("file buffer is full")
)

type File struct {
	size int64
	file *os.File
}

func CreateFile(filename string) (*File, error) {
	var f, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}

	return &File{
		size: 0,
		file: f,
	}, nil
}

func OpenFile(filename string) (*File, error) {
	var f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &File{
		size: stat.Size(),
		file: f,
	}, nil
}

func (this *File) Size() int64 {
	return this.size
}

func (this *File) Write(p []byte) (n int, err error) {
	n, err = this.file.Write(p)
	this.size += int64(n)
	return n, err
}

func (this *File) Close() (err error) {
	if this.file != nil {
		err = this.file.Close()
	}
	this.size = 0
	return err
}

type FileWriter2 struct {
	level      Level
	dir        string
	filename   string
	maxSize    int64
	maxAge     int64
	bufferSize int

	file       *File
	pool       *sync.Pool
	bufferChan chan *bytes.Buffer
	closed     int32
	wg         *sync.WaitGroup
}

func NewFileWriter2(level Level, opts ...FileWriterOption) *FileWriter2 {
	var w = &FileWriter2{}
	w.level = level
	w.dir = kLogDir
	w.maxSize = 10 * 1024 * 1024
	w.maxAge = 0
	w.bufferSize = 1024 * 100

	//for _, opt := range opts {
	//	opt.Apply(w)
	//}

	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return nil
	}
	w.pool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
	w.bufferChan = make(chan *bytes.Buffer, w.bufferSize)
	w.filename = path.Join(w.dir, kLogFile)
	w.closed = 0
	w.wg = &sync.WaitGroup{}
	w.wg.Add(1)

	go w.daemon()

	return w
}

func (this *FileWriter2) getBuffer() *bytes.Buffer {
	return this.pool.Get().(*bytes.Buffer)
}

func (this *FileWriter2) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	this.pool.Put(buf)
}

func (this *FileWriter2) Level() Level {
	return this.level
}

func (this *FileWriter2) WriteMessage(logId, service, instance, prefix, logTime string, level Level, file, line, msg string) {
	if atomic.LoadInt32(&this.closed) == 1 {
		return
	}
	var buf = this.getBuffer()
	buf.WriteByte('[')
	buf.WriteString(logId)
	buf.WriteByte(']')
	buf.WriteByte(' ')
	buf.WriteString(service)
	buf.WriteString(instance)
	buf.WriteString(prefix)
	buf.WriteString(logTime)
	buf.WriteByte(' ')
	buf.WriteString(LevelNames[level])
	buf.WriteByte(' ')
	buf.WriteString(file)
	buf.WriteString(":")
	buf.WriteString(line)
	buf.WriteString(" ")
	buf.WriteString(msg)

	this.bufferChan <- buf

	//select {
	//case this.bufferChan <- buf:
	//	return
	//default:
	//	this.putBuffer(buf)
	//}
}

//
//func (this *FileWriter2) Write(p []byte) (int, error) {
//	if atomic.LoadInt32(&this.closed) == 1 {
//		return 0, ErrFileClosed
//	}
//	var buf = this.getBuffer()
//	buf.Write(p)
//
//	//this.bufferChan <- buf
//
//	select {
//	case this.bufferChan <- buf:
//		return len(p), nil
//	default:
//		this.putBuffer(buf)
//		return 0, ErrFileBufferFull
//	}
//}

func (this *FileWriter2) Close() error {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) == false {
		this.wg.Wait()
		return nil
	}
	close(this.bufferChan)
	this.wg.Wait()
	return this.close()
}

func (this *FileWriter2) close() error {
	var err error
	if this.file != nil {
		err = this.file.Close()
	}
	this.file = nil
	return err
}

func (this *FileWriter2) daemon() {
	for {
		select {
		case buf, ok := <-this.bufferChan:
			if ok {
				this.write(buf.Bytes())
				this.putBuffer(buf)
			}
		}

		if atomic.LoadInt32(&this.closed) == 0 {
			continue
		}

		for buf := range this.bufferChan {
			this.write(buf.Bytes())
			this.putBuffer(buf)
		}
		break
	}
	this.wg.Done()
}

func (this *FileWriter2) write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	pLen := int64(len(p))
	if this.file == nil {
		if err = this.openOrCreate(pLen); err != nil {
			return 0, err
		}
	}

	if this.file.size+pLen >= this.maxSize {
		if err := this.rotate(); err != nil {
			return 0, err
		}
	}

	return this.file.Write(p)
}

func (this *FileWriter2) openOrCreate(pLen int64) error {
	this.clean()

	// 获取log文件信息
	info, err := os.Stat(this.filename)
	if os.IsNotExist(err) {
		// 如果log文件不存在，直接创建新的log文件
		return this.create()
	}
	if err != nil {
		return err
	}

	// 文件存在，但是其文件大小已超出设定的阈值
	if info.Size()+pLen >= this.maxSize {
		return this.rotate()
	}

	// 打开现有的文件
	file, err := OpenFile(this.filename)
	if err != nil {
		// 如果打开文件出错，则创建新的文件
		return this.create()
	}
	this.file = file
	return nil
}

func (this *FileWriter2) create() error {
	file, err := CreateFile(this.filename)
	if err != nil {
		return err
	}
	this.file = file
	return nil
}

func (this *FileWriter2) rename() error {
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

func (this *FileWriter2) rotate() error {
	if err := this.close(); err != nil {
		return err
	}

	if err := this.rename(); err != nil {
		return err
	}

	if err := this.create(); err != nil {
		return err
	}
	this.clean()
	return nil
}

func (this *FileWriter2) clean() {
	if this.maxAge <= 0 {
		return
	}
	this.wg.Add(1)
	go func() {
		var dir = filepath.Dir(this.dir)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) (rErr error) {
			defer func() {
				if r := recover(); r != nil {
				}
			}()

			if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-this.maxAge) {
				if filepath.Ext(info.Name()) == kLogFileExt && info.Name() != kLogFile {
					rErr = os.Remove(path)
				}
			}
			return rErr
		})
		this.wg.Done()
	}()
}
