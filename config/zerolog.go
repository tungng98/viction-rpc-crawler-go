package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitZerolog(configDir string, level, consoleLevel int8) *os.File {
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{Writer: zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.DateTime}},
		Level:  zerolog.Level(consoleLevel),
	}

	logFile := ""
	if configDir != "" {
		date := time.Now().UTC().Format("20060102")
		logFile = path.Join(configDir, "logs", fmt.Sprintf("viction-rcp-crawler-%s.log", date))
	}
	logDir := path.Join(configDir, "logs")
	if !isExist(logDir) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
			log.Err(err).Msgf("Cannot create log file: %s", logFile)
			return nil
		}
	}
	if level < 0 {
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
		return nil
	}
	fileStream, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	fileWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{Writer: fileStream},
		Level:  zerolog.Level(level),
	}
	if err != nil {
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
		log.Err(err).Msgf("Cannot create log file: %s", logFile)
		return nil
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
	return fileStream
}
