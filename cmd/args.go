package cmd

type Args struct {
	IndexBlockTx *IndexBlockTxCmd `arg:"subcommand:index" help:"Index BlockHashes and TxHashes and stored in MongoDB"`
	Service      *ServiceCmd      `arg:"subcommand:service" help:"Run in service mode using predefined configuration"`
}

type IndexBlockTxCmd struct {
	StartBlock  uint64 `arg:"-f,--from" help:"Block number to start the crawling process"`
	EndBlock    uint64 `arg:"-t,--to" help:"Block number to stop the crawling process"`
	WorkerCount int    `arg:"--worker" help:"Number of concurrent routine to fetch blocks from RPC"`
	BatchSize   int    `arg:"--batch" help:"Number of blocks to persisted in one write operation"`
	Forced      bool   `arg:"--force" help:"Ignore the checkpoint number stored in database"`
	IncludeTxs  bool   `arg:"--txs" help:"Save transaction data to database"`
	PostgreSQL  string `arg:"--pgsql" help:"Connection string to PostgreSQL"`
	RpcUrl      string `arg:"--rpc" help:"Blockchain node RPC endpoint"`
}

type ServiceCmd struct {
	Configuration string `arg:"-c,--config" help:"Configuration file contains services definition"`
}
