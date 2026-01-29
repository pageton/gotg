package log

import (
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
