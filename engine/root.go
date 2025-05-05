package engine

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:     "viction-crawler",
		Short:   "Viction Blockchain data crawler.",
		Version: version(),
	}
	rootCmd.AddCommand(BenchmarkCmd())
	rootCmd.AddCommand(DatabaseCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
