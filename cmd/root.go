package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/okaris/tygo/config"
	"github.com/okaris/tygo/tygo"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "tygo",
		Short: "Tool for generating Typescript from Go types",
		Long:  `Tygo generates Typescript interfaces and constants from Go files by parsing them.`,
	}

	rootCmd.PersistentFlags().
		String("config", "tygo.yaml", "config file to load (default is tygo.yaml in the current folder)")
	rootCmd.Version = FullVersion()
	rootCmd.PersistentFlags().BoolP("debug", "D", false, "Debug mode (prints debug messages)")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate and write to disk",
		Run:   generate,
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generate(cmd *cobra.Command, args []string) {
	cfgFilepath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatal(err)
	}
	tygoConfig := config.ReadFromFilepath(cfgFilepath)
	t := tygo.New(&tygoConfig)

	err = t.Generate()
	if err != nil {
		log.Fatalf("Tygo failed: %v", err)
	}
}
