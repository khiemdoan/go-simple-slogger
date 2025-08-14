// Author: Khiem Doan
// Github: https://github.com/khiemdoan
// Email: doankhiem.crazy<at>gmail.com

package slogger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DeRuina/timberjack"
	"gitlab.com/greyxor/slogor"
)

type pack struct {
	handler slog.Handler
	record  slog.Record
}

var (
	pipe chan pack
)

func init() {
	pipe = make(chan pack)
	go asyncHandle()
}

func asyncHandle() {
	for pack := range pipe {
		if err := handle(pack.handler, pack.record); err != nil {
			slog.Error("Failed to handle log record", slog.String("error", err.Error()))
		}
	}
}

func handle(h slog.Handler, r slog.Record) error {
	ctx := context.Background()
	if !h.Enabled(ctx, r.Level) {
		return nil
	}
	return h.Handle(ctx, r)
}

type option struct {
	dir        string
	file       string
	maxSize    int
	maxBackups int
	maxAge     int
}

type optionFunc func(*option)

type slogger struct {
	handlers []slog.Handler
	async    bool
}

func WithDir(dir string) optionFunc {
	return func(o *option) {
		o.dir = dir
	}
}

func WithFile(file string) optionFunc {
	return func(o *option) {
		o.file = file
	}
}

func WithMaxSize(maxSize int) optionFunc {
	return func(o *option) {
		o.maxSize = maxSize
	}
}

func WithMaxBackups(maxBackups int) optionFunc {
	return func(o *option) {
		o.maxBackups = maxBackups
	}
}

func WithMaxAge(maxAge int) optionFunc {
	return func(o *option) {
		o.maxAge = maxAge
	}
}

func NewSlogger(opts ...optionFunc) *slogger {
	return &slogger{
		handlers: createHandlers(opts...),
		async:    false,
	}
}

func NewAsyncSlogger(opts ...optionFunc) *slogger {
	return &slogger{
		handlers: createHandlers(opts...),
		async:    true,
	}
}

func createHandlers(opts ...optionFunc) []slog.Handler {
	option := &option{ // default options
		dir:        "logs",
		file:       "app.log",
		maxSize:    10,
		maxBackups: 3,
		maxAge:     7,
	}

	for _, opt := range opts {
		opt(option)
	}

	if fileInfo, err := os.Stat(option.dir); os.IsNotExist(err) {
		os.MkdirAll(option.dir, 0755)
	} else {
		if !fileInfo.IsDir() {
			os.Remove(option.dir)
			os.MkdirAll(option.dir, 0755)
		}
	}

	fileHandler := slog.NewJSONHandler(&timberjack.Logger{
		Filename:   filepath.Join(option.dir, option.file),
		MaxSize:    option.maxSize,
		MaxBackups: option.maxBackups,
		MaxAge:     option.maxAge,
	}, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelWarn,
	})

	consoleHandler := slogor.NewHandler(
		os.Stdout,
		slogor.SetLevel(slog.LevelDebug),
		slogor.SetTimeFormat(time.DateTime),
		slogor.ShowSource(),
	)

	return []slog.Handler{fileHandler, consoleHandler}
}

func (l *slogger) log(level slog.Level, msg string, attrs ...slog.Attr) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	record := slog.NewRecord(time.Now(), level, msg, pcs[0])
	record.AddAttrs(attrs...)

	for _, handler := range l.handlers {
		if l.async {
			pipe <- pack{handler, record}
		} else {
			if err := handle(handler, record); err != nil {
				slog.Error("Failed to handle log record", slog.String("error", err.Error()))
			}
		}
	}
}

func (l *slogger) Debug(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelDebug, msg, attrs...)
}

func (l *slogger) Info(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelInfo, msg, attrs...)
}

func (l *slogger) Warn(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelWarn, msg, attrs...)
}

func (l *slogger) Error(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelError, msg, attrs...)
}
