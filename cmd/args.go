package cmd

type Args struct {
	IndexBlockTx      *IndexBlockTxCmd      `arg:"subcommand:index" help:"Index BlockHashes and TxHashes and stored in MongoDB"`
	ManageDatabase    *DatabaseCmd          `arg:"subcommand:database" help:""`
	ScanBlockForError *ScanBlockForErrorCmd `arg:"subcommand:scan" help:"Scan blocks to find problematic BlockHash that caused issue"`
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
}

type ScanBlockForErrorCmd struct {
	StartBlock   uint64 `arg:"-f,--from" help:"Block number to start the crawling process"`
	EndBlock     uint64 `arg:"-t,--to" help:"Block number to stop the crawling process"`
	WorkerCount  int32  `arg:"--worker" help:"Number of concurrent routine to fetch blocks from RPC"`
	BatchSize    int    `arg:"--batch" help:"Number of blocks to persisted in one write operation"`
	NoCheckpoint bool   `arg:"--no-cp" help:"Ignore the checkpoint number stored in database"`
	NoSaveTrace  bool   `arg:"--no-save" help:"Don't save trace result into database"`
}
