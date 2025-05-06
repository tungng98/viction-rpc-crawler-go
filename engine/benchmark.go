package engine

import (
	"math/big"
	"viction-rpc-crawler-go/config"
	"viction-rpc-crawler-go/rpc"
	"viction-rpc-crawler-go/svc"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/multiplex"
)

type BenchmarkModule struct {
	config *config.RootConfig
	logger zerolog.Logger
}

func NewBenchmarkModule(c *Controller, cmdName string) *BenchmarkModule {
	return &BenchmarkModule{
		config: c.Root,
		logger: c.CommandLogger("benchmark", cmdName),
	}
}

func (m *BenchmarkModule) GetBlocks(from, to *big.Int, batchSize int) error {
	m.logger.Info().Msg("Start eth_getBlockByNumber benchmark.")
	rpcClient, err := rpc.Connect(m.config.Blockchain.RpcUrl)
	if err != nil {
		return err
	}
	c := svc.NewController(m.config, nil, rpcClient, config.NewZerologLogger(m.logger))
	go c.DispatchOnce("GetBlocks", "get_blocks_range", multiplex.ExecParams{
		"from_block_number": from,
		"to_block_number":   to,
		"batch_size":        batchSize,
	})
	c.Run()
	return nil
}

func (m *BenchmarkModule) TraceBlocks(from, to *big.Int, batchSize int) error {
	m.logger.Info().Msg("Start debug_traceBlockByNumber benchmark.")
	rpcClient, err := rpc.Connect(m.config.Blockchain.RpcUrl)
	if err != nil {
		return err
	}
	c := svc.NewController(m.config, nil, rpcClient, config.NewZerologLogger(m.logger))
	go c.DispatchOnce("TraceBlocks", "trace_blocks_range", multiplex.ExecParams{
		"from_block_number": from,
		"to_block_number":   to,
		"batch_size":        batchSize,
	})
	c.Run()
	return nil
}

func (m *BenchmarkModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

func BenchmarkCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Measure performance of Viction node.",
	}

	getBlockCmd := &cobra.Command{
		Use:   "get-block",
		Short: "Benchmark using eth_getBlockByNumber method.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseBenchmarkFlags(cmd)
			c.ConfigFromCli(flags.Configs)
			m := NewBenchmarkModule(c, "getBlock")
			m.logError(m.GetBlocks(flags.From, flags.To, flags.Batch))
		},
	}
	getBlockCmd.Flags().Int("batch", 900, "Batch size.")
	getBlockCmd.Flags().Uint64P("from", "f", 0, "Start block number.")
	getBlockCmd.Flags().String("rpc", "", "RPC URL.")
	getBlockCmd.Flags().Uint64P("to", "t", 1000, "To block number.")
	rootCmd.AddCommand(getBlockCmd)

	traceBlocksCmd := &cobra.Command{
		Use:   "trace-block",
		Short: "Benchmark using debug_traceBlockByNumber method.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseBenchmarkFlags(cmd)
			c.ConfigFromCli(flags.Configs)
			m := NewBenchmarkModule(c, "traceBlock")
			m.logError(m.TraceBlocks(flags.From, flags.To, flags.Batch))
		},
	}
	traceBlocksCmd.Flags().Int("batch", 900, "Batch size.")
	traceBlocksCmd.Flags().Uint64P("from", "f", 0, "Start block number.")
	traceBlocksCmd.Flags().String("rpc", "", "RPC URL.")
	traceBlocksCmd.Flags().Uint64P("to", "t", 1000, "To block number.")
	rootCmd.AddCommand(traceBlocksCmd)

	return rootCmd
}

type BenchmarkFlags struct {
	Batch int
	From  *big.Int
	To    *big.Int

	Configs map[string]interface{}
}

func ParseBenchmarkFlags(cmd *cobra.Command) *BenchmarkFlags {
	batch, _ := cmd.Flags().GetInt("batch")
	from, _ := cmd.Flags().GetUint64("from")
	rpcUrl, _ := cmd.Flags().GetString("rpc")
	to, _ := cmd.Flags().GetUint64("to")

	configs := make(map[string]interface{})
	if rpcUrl != "" {
		configs[config.BlockchainRpcUrlKey] = rpcUrl
	}

	return &BenchmarkFlags{
		Batch:   batch,
		From:    new(big.Int).SetUint64(from),
		To:      new(big.Int).SetUint64(to),
		Configs: configs,
	}
}
