package logging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"consul-journey/internal"
	"consul-journey/internal/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(config *Config) (l *zap.Logger, err error) {
	cCore := newConsoleCore(config.Console)
	fCore, err := newFileCore(config.File)
	if err != nil {
		return nil, fmt.Errorf("failed to create file core: %w", err)
	}

	// Chain the cores
	var core zapcore.Core
	if fCore != nil && cCore != nil {
		core = zapcore.NewTee(cCore, fCore)
	} else if fCore != nil {
		core = fCore
	} else if cCore != nil {
		core = cCore
	} else {
		return zap.NewNop(), nil
	}

	// Create the logger
	l = zap.New(
		core,
		zap.WithCaller(false),
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
	)
	if config.Version {
		l = l.With(zap.String("ver", internal.Version()))
	}
	if config.Revision {
		l = l.With(zap.String("rev", internal.Revision()))
	}
	return l, nil
}

func newConsoleCore(cfg *ConsoleConfig) zapcore.Core {
	if !cfg.Enabled {
		return nil
	}
	switch cfg.Format {
	case LoggingFormatJSON:
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(newJsonEncoderConfig()),
			zapcore.AddSync(os.Stderr),
			cfg.Level.zapLevel(),
		)
	}
	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(newColoredTextEncoderConfig()),
		zapcore.AddSync(os.Stderr),
		cfg.Level.zapLevel(),
	)
}

func newFileCore(cfg *FileConfig) (zapcore.Core, error) {
	if err := setupAndValidateLogsFolder(cfg.Folder); err != nil {
		return nil, err
	}
	syncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Folder, cfg.Name+".log"),
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	})
	switch cfg.Format {
	case LoggingFormatText:
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(newTextEncoderConfig()),
			syncer,
			cfg.Level.zapLevel(),
		), nil
	}
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(newJsonEncoderConfig()),
		syncer,
		cfg.Level.zapLevel(),
	), nil
}

func newColoredTextEncoderConfig() zapcore.EncoderConfig {
	cfg := newTextEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg
}

func newTextEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.999999"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func newJsonEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.999999Z"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func setupAndValidateLogsFolder(path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("failed to create logs folder: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get logs folder info: %w", err)
	}
	if !info.IsDir() {
		return errors.New("logs folder is not a directory")
	}

	testFile := filepath.Join(path, "."+utils.RandomString(16))
	if err := os.WriteFile(testFile, []byte{}, 0o600); err != nil {
		return fmt.Errorf("logs folder is not writable: %w", err)
	}
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to clean up test file: %w", err)
	}
	return nil
}
