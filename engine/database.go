package engine

import (
	"viction-rpc-crawler-go/config"
	"viction-rpc-crawler-go/db"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type DatabaseModule struct {
	config *config.RootConfig
	logger zerolog.Logger
}

func NewDatabaseModule(c *Controller, cmdName string) *DatabaseModule {
	return &DatabaseModule{
		config: c.Root,
		logger: c.CommandLogger("database", cmdName),
	}
}

func (m *DatabaseModule) Migrate() error {
	c, err := db.Connect(m.config.Database.PostgreSQL, "")
	if err != nil {
		return err
	}
	err = c.Migrate()
	if err != nil {
		return err
	}

	log.Info().Msg("Migration successful!")
	return nil
}

func (m *DatabaseModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

func DatabaseCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "database",
		Short: "Perform database maintenance.",
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Pre-populate tables for working this tool.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseDatabaseFlags(cmd)
			c.ConfigFromCli(flags.Configs)
			m := NewDatabaseModule(c, "migrate")
			m.logError(m.Migrate())
		},
	}
	rootCmd.AddCommand(migrateCmd)

	return rootCmd
}

type DatabaseFlags struct {
	Configs map[string]interface{}
}

func ParseDatabaseFlags(cmd *cobra.Command) *DatabaseFlags {
	configs := make(map[string]interface{})

	return &DatabaseFlags{
		Configs: configs,
	}
}
