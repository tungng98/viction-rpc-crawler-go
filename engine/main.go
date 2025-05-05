package engine

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"viction-rpc-crawler-go/filesystem"

	"github.com/tforce-io/tf-golib/opx"
)

var majorVersion = 0
var minorVersion = 1
var patchVersion = 0
var gitCommit, gitDate, gitBranch string

func version() string {
	originDate := time.Date(2024, time.August, 13, 0, 0, 0, 0, time.UTC)
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(gitDate == "", time.Now().UTC(), gitDate2)
	duration := buildDate.Sub(originDate)
	minor := minorVersion
	patch := strconv.Itoa(patchVersion)
	if gitBranch == "master" {
		// do nothing
	} else if gitBranch == "release" {
		minor += 1
		patch = patch + "-rc"
	} else if strings.Contains(gitBranch, "feat/") {
		minor += 1
		patch = patch + "-dev"
	} else {
		patch = strconv.Itoa(patchVersion+1) + "-dev"
	}
	if gitCommit != "" {
		return fmt.Sprintf("%d.%d.%s.%d-%s", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000), gitCommit)
	}
	return fmt.Sprintf("%d.%d.%s.%d", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000))
}

func InitApp() *Controller {
	cfg := NewController(true)

	pwd, _ := os.Getwd()
	pwd, _ = filesystem.GetAbsPath(pwd)
	exec, _ := os.Executable()
	exec, _ = filesystem.GetAbsPath(exec)

	cfg.Logger.Info().Msgf("TF UNIFILER v%s", version())
	cfg.Logger.Info().Msg("RPC CRAWLER")
	cfg.Logger.Info().Msgf("Working directory: %s", pwd)
	cfg.Logger.Info().Msgf("Config directory: %s", cfg.Root.ConfigDir)
	cfg.Logger.Info().Msgf("Config file: %s", cfg.Root.ConfigFile)
	cfg.Logger.Info().Msgf("Executable file: %s", exec)
	cfg.Logger.Info().Msgf("Portable mode: %t", cfg.Root.IsPortable)
	cfg.Logger.Info().Msg("-----------------")

	return cfg
}
