package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
	"viction-rpc-crawler-go/filesystem"

	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/opx"
)

func InitZerolog(configDir string, useFS bool) (zerolog.Logger, *os.File, error) {
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.DateTime},
		},
		Level: zerolog.TraceLevel,
	}

	logFile, err := InitLogFile(useFS, configDir)
	if logFile == nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		return consoleLogger, nil, err
	}

	fileWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: logFile,
		},
		Level: zerolog.TraceLevel,
	}
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	return logger, logFile, nil
}

func InitLogFile(useFS bool, workdingDir string) (*os.File, error) {
	if !useFS {
		return nil, nil
	}
	logDir := path.Join(opx.Ternary(workdingDir == "", ".", workdingDir), "logs")
	if !filesystem.IsExist(logDir) {
		err := filesystem.CreateDirectoryRecursive(logDir)
		if err != nil {
			return nil, err
		}
	}
	execPath := ExecPath()
	logFileName := fmt.Sprintf(execPath.Name.Name+"-%s.log", time.Now().UTC().Format("20060102"))
	logFilePath := filepath.Join(logDir, logFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

type ZerologLogger struct {
	i *ZerologLoggerInternal
}

type ZerologLoggerInternal struct {
	Logger zerolog.Logger
}

func NewZerologLogger(logger zerolog.Logger) *ZerologLogger {
	return &ZerologLogger{
		i: &ZerologLoggerInternal{
			Logger: logger,
		},
	}
}

func (l ZerologLogger) Error(err error, v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.i.Logger.Error().Err(err).Msg(msg)
}

func (l ZerologLogger) Errorf(err error, format string, v ...interface{}) {
	l.i.Logger.Error().Err(err).Msgf(format, v...)
}

func (l ZerologLogger) Warn(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.i.Logger.Warn().Msg(msg)
}

// Print a message with Warn level with format.
func (l ZerologLogger) Warnf(format string, v ...interface{}) {
	l.i.Logger.Warn().Msgf(format, v...)
}

// Print a message with Info level.
func (l ZerologLogger) Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.i.Logger.Info().Msg(msg)
}

// Print a message with Info level with format.
func (l ZerologLogger) Infof(format string, v ...interface{}) {
	l.i.Logger.Info().Msgf(format, v...)
}

// Print a message with Debug level.
func (l ZerologLogger) Debug(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.i.Logger.Debug().Msg(msg)
}

// Print a message with Debug level with format.
func (l ZerologLogger) Debugf(format string, v ...interface{}) {
	l.i.Logger.Debug().Msgf(format, v...)
}

// Print a message with Trace level.
func (l ZerologLogger) Trace(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.i.Logger.Trace().Msg(msg)
}

// Print a message with Trace level with format.
func (l ZerologLogger) Tracef(format string, v ...interface{}) {
	l.i.Logger.Trace().Msgf(format, v...)
}
