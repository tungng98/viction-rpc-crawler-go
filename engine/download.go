package engine

import (
	"math/big"
	"viction-rpc-crawler-go/config"
	"viction-rpc-crawler-go/rpc"
	"viction-rpc-crawler-go/svc"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/multiplex"
	"github.com/tforce-io/tf-golib/opx"
)

type DownloadModule struct {
	config *config.RootConfig
	logger zerolog.Logger
}

func NewDownloadModule(c *Controller, cmdName string) *DownloadModule {
	return &DownloadModule{
		config: c.Root,
		logger: c.CommandLogger("download", cmdName),
	}
}

func (m *DownloadModule) GetBlocks(from, to *big.Int, batchSize int, root string) error {
	m.logger.Info().Msg("Start eth_getBlockByNumber download.")
	rpcClient, err := rpc.Connect(m.config.Blockchain.RpcUrl)
	if err != nil {
		return err
	}
	c := svc.NewController(m.config, nil, rpcClient, config.NewZerologLogger(m.logger))
	go c.DispatchOnce("IndexBlock", "download", multiplex.ExecParams{
		"from_block_number": from,
		"to_block_number":   to,
		"batch_size":        batchSize,
		"root":              opx.Ternary(root == "", m.config.FileSystem.RootPath, root),
	})
	c.Run()
	return nil
}

func (m *DownloadModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

func DownloadCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "download",
		Short: "Download response from RPC and store to filesystem.",
	}

	getBlockCmd := &cobra.Command{
		Use:   "get-block",
		Short: "Download eth_getBlockByNumber data.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseDownloadFlags(cmd)
			c.ConfigFromCli(flags.Configs)
			m := NewDownloadModule(c, "getBlock")
			m.logError(m.GetBlocks(flags.From, flags.To, flags.Batch, flags.Root))
		},
	}
	getBlockCmd.Flags().Int("batch", 1, "Batch size.")
	getBlockCmd.Flags().Uint64P("from", "f", 1, "Start block number.")
	getBlockCmd.Flags().String("rpc", "", "RPC URL.")
	getBlockCmd.Flags().String("root", "", "Root output dir.")
	getBlockCmd.Flags().Uint64P("to", "t", 1, "To block number.")
	rootCmd.AddCommand(getBlockCmd)

	return rootCmd
}

type DownloadFlags struct {
	Batch int
	From  *big.Int
	Root  string
	To    *big.Int

	Configs map[string]interface{}
}

func ParseDownloadFlags(cmd *cobra.Command) *DownloadFlags {
	batch, _ := cmd.Flags().GetInt("batch")
	from, _ := cmd.Flags().GetUint64("from")
	rootDir, _ := cmd.Flags().GetString("root")
	rpcUrl, _ := cmd.Flags().GetString("rpc")
	to, _ := cmd.Flags().GetUint64("to")

	configs := make(map[string]interface{})
	if rpcUrl != "" {
		configs[config.BlockchainRpcUrlKey] = rpcUrl
	}

	return &DownloadFlags{
		Batch:   batch,
		From:    new(big.Int).SetUint64(from),
		Root:    rootDir,
		To:      new(big.Int).SetUint64(to),
		Configs: configs,
	}
}
