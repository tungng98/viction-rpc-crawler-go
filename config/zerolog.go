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
