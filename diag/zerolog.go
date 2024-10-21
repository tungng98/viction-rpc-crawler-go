package diag

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitZerolog(configDir string) *os.File {
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{Writer: zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.DateTime}},
		Level:  zerolog.DebugLevel,
	}

	logFile := ""
	if configDir != "" {
		date := time.Now().UTC().Format("20060102")
		logFile = path.Join(configDir, "logs", fmt.Sprintf("unifiler-%s.log", date))
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
	fileWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
		log.Err(err).Msgf("Cannot create log file: %s", logFile)
		return nil
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
	return fileWriter
}

func isExist(fPath string) bool {
	_, err := os.Stat(fPath)
	return !os.IsNotExist(err)
}
