package cmd

type Args struct {
	IndexBlockTx   *IndexBlockTxCmd `arg:"subcommand:index" help:"Index BlockHashes and TxHashes and stored in MongoDB"`
	ManageDatabase *DatabaseCmd     `arg:"subcommand:database" help:""`
}

type DatabaseCmd struct {
	Migrate *DatabaseMigrateCmd `arg:"subcommand:migrate" help:"Pre-populate tables for working this tool"`
}

type DatabaseMigrateCmd struct {
	PostgreSQL string `arg:"--pgsql" help:"Connection string to PostgreSQL"`
}

type IndexBlockTxCmd struct {
	StartBlock  uint64 `arg:"-f,--from" help:"Block number to start the crawling process"`
	EndBlock    uint64 `arg:"-t,--to" help:"Block number to stop the crawling process"`
	WorkerCount int    `arg:"--worker" help:"Number of concurrent routine to fetch blocks from RPC"`
	BatchSize   int    `arg:"--batch" help:"Number of blocks to persisted in one write operation"`
	Forced      bool   `arg:"--force" help:"Ignore the checkpoint number stored in database"`
	PostgreSQL  string `arg:"--pgsql" help:"Connection string to PostgreSQL"`
	RpcUrl      string `arg:"--rpc" help:"Blockchain node RPC endpoint"`
}
