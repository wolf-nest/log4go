package log4go

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrFileClosed     = errors.New("file already closed")
	ErrFileBufferFull = errors.New("file buffer is full")
)

const (
	kLogDir       = "./logs"
	kLogFile      = "temp_log.log"
	kLogFileExt   = ".log"
	kBuffChanSize = 1024 * 10
	kMaxFileSize  = 1024 * 1024 * 10
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
		if sec <= 0 {
			return
		}
		w.maxAge = sec
	})
}

func WithMaxSize(mb int64) FileWriterOption {
	return fwOptionFunc(func(w *FileWriter) {
		if mb <= 0 {
			return
		}
		w.maxSize = mb * 1024 * 1024
	})
}

func WithLogDir(dir string) FileWriterOption {
	return fwOptionFunc(func(w *FileWriter) {
		if strings.TrimSpace(dir) == "" {
			return
		}
		w.dir = dir
	})
}

func WithBuffChanSize(size int) FileWriterOption {
	return fwOptionFunc(func(w *FileWriter) {
		if size <= 0 {
			size = kBuffChanSize
		}
		w.buffChanSize = size
	})
}

type File struct {
	size int64
	file *os.File
}

func CreateFile(filename string) (*File, error) {
	var f, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &File{
		size: 0,
		file: f,
	}, nil
}

func OpenFile(filename string) (*File, error) {
	var f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
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

type FileWriter struct {
	level        Level
	dir          string
	filename     string
	maxSize      int64
	maxAge       int64
	buffChanSize int

	file     *File
	pool     *sync.Pool
	buffChan chan *bytes.Buffer
	sync     chan struct{}
	syncWg   sync.WaitGroup
	closed   int32
	wg       sync.WaitGroup
}

func NewFileWriter2(level Level, opts ...FileWriterOption) *FileWriter {
	var w = &FileWriter{}
	w.level = level
	w.dir = kLogDir
	w.maxSize = kMaxFileSize
	w.maxAge = 0
	w.buffChanSize = kBuffChanSize

	for _, opt := range opts {
		opt.Apply(w)
	}

	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return nil
	}
	w.pool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
	w.buffChan = make(chan *bytes.Buffer, w.buffChanSize)
	w.sync = make(chan struct{}, 1)
	w.filename = path.Join(w.dir, kLogFile)
	w.closed = 0
	w.wg.Add(1)
	go w.daemon()
	return w
}

func (this *FileWriter) getBuffer() *bytes.Buffer {
	return this.pool.Get().(*bytes.Buffer)
}

func (this *FileWriter) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	this.pool.Put(buf)
}

func (this *FileWriter) Close() error {
	if old := atomic.SwapInt32(&this.closed, 1); old != 0 {
		this.wg.Wait()
		return nil
	}
	close(this.buffChan)
	this.wg.Wait()
	return this.close()
}

func (this *FileWriter) close() error {
	var err error
	if this.file != nil {
		err = this.file.Close()
	}
	this.file = nil
	return err
}

func (this *FileWriter) Level() Level {
	return this.level
}

func (this *FileWriter) Sync() error {
	if atomic.LoadInt32(&this.closed) == 1 {
		return ErrFileClosed
	}
	this.syncWg.Add(1)
	this.sync <- struct{}{}
	this.syncWg.Wait()
	return nil
}

func (this *FileWriter) WriteMessage(logId, service, instance, prefix, logTime string, level Level, file, line, msg string) {
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

	this.syncWg.Wait()
	this.buffChan <- buf

	//select {
	//case this.buffChan <- buf:
	//	return
	//default:
	//	this.putBuffer(buf)
	//}
}

func (this *FileWriter) Write(p []byte) (int, error) {
	if atomic.LoadInt32(&this.closed) == 1 {
		return 0, ErrFileClosed
	}
	var buf = this.getBuffer()
	buf.Write(p)

	this.syncWg.Wait()
	this.buffChan <- buf
	return len(p), nil

	//select {
	//case this.buffChan <- buf:
	//	return len(p), nil
	//default:
	//	this.putBuffer(buf)
	//	return 0, ErrFileBufferFull
	//}
}

func (this *FileWriter) daemon() {
	defer this.wg.Done()
	for {
		select {
		case <-this.sync:
			close(this.buffChan)
			for buf := range this.buffChan {
				this.write(buf.Bytes())
				this.putBuffer(buf)
			}
			this.buffChan = make(chan *bytes.Buffer, this.buffChanSize)
			this.syncWg.Done()
		default:
			select {
			case buf, ok := <-this.buffChan:
				if ok {
					this.write(buf.Bytes())
					this.putBuffer(buf)
				}
			}

			if atomic.LoadInt32(&this.closed) == 0 {
				continue
			}

			for buf := range this.buffChan {
				this.write(buf.Bytes())
				this.putBuffer(buf)
			}
			return
		}
	}
}

func (this *FileWriter) write(p []byte) (n int, err error) {
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

func (this *FileWriter) openOrCreate(pLen int64) error {
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

func (this *FileWriter) create() error {
	file, err := CreateFile(this.filename)
	if err != nil {
		return err
	}
	this.file = file
	return nil
}

func (this *FileWriter) rename() error {
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

	if err := this.rename(); err != nil {
		return err
	}

	if err := this.create(); err != nil {
		return err
	}
	this.clean()
	return nil
}

func (this *FileWriter) clean() {
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
