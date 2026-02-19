package log

import (
	"fmt"
	"os"
	"sync"
)

type Writer interface {
	Write(data []byte) error
	Close() error
}

type ConsoleWriter struct {
	mu  sync.Mutex
	out *os.File
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{out: os.Stderr}
}

func (w *ConsoleWriter) Write(data []byte) error {
	w.mu.Lock()
	_, err := w.out.Write(data)
	w.mu.Unlock()
	return err
}

func (w *ConsoleWriter) Close() error { return nil }

type FileWriter struct {
	mu   sync.Mutex
	file *os.File
}

func NewFileWriter(path string) (*FileWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileWriter{file: f}, nil
}

func (w *FileWriter) Write(data []byte) error {
	w.mu.Lock()
	_, err := w.file.Write(data)
	w.mu.Unlock()
	return err
}

func (w *FileWriter) Close() error {
	return w.file.Close()
}

type MultiWriter struct {
	writers []Writer
}

func NewMultiWriter(writers ...Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (mw *MultiWriter) Write(data []byte) error {
	for _, w := range mw.writers {
		if err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (mw *MultiWriter) Close() error {
	for _, w := range mw.writers {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

// RotatingFileWriter wraps file writing with automatic size-based rotation.
type RotatingFileWriter struct {
	mu         sync.Mutex
	path       string
	file       *os.File
	size       int64
	maxSize    int64
	maxBackups int
}

const (
	defaultMaxSizeBytes = 100 * 1024 * 1024 // 100 MB
	defaultMaxBackups   = 3
)

func NewRotatingFileWriter(path string, maxSize int64, maxBackups int) (*RotatingFileWriter, error) {
	if maxSize <= 0 {
		maxSize = defaultMaxSizeBytes
	}
	if maxBackups <= 0 {
		maxBackups = defaultMaxBackups
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	return &RotatingFileWriter{
		path:       path,
		file:       f,
		size:       info.Size(),
		maxSize:    maxSize,
		maxBackups: maxBackups,
	}, nil
}

func (w *RotatingFileWriter) Write(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.size+int64(len(data)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return err
		}
	}

	n, err := w.file.Write(data)
	w.size += int64(n)
	return err
}

func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

func (w *RotatingFileWriter) rotate() error {
	w.file.Close()

	// Shift existing backups: .3 -> delete, .2 -> .3, .1 -> .2
	for i := w.maxBackups; i >= 1; i-- {
		src := fmt.Sprintf("%s.%d.log", w.path, i)
		if i == w.maxBackups {
			os.Remove(src)
		} else {
			dst := fmt.Sprintf("%s.%d.log", w.path, i+1)
			os.Rename(src, dst)
		}
	}

	// Current file -> .1
	os.Rename(w.path, fmt.Sprintf("%s.%d.log", w.path, 1))

	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.size = 0
	return nil
}
