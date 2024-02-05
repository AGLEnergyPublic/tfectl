package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// The log flag value.
var l string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "tfectl",
	Short:             "Query TFE and TFC from the command line.",
	Long:              `Query TFE and TFC from the command line.`,
	Version:           "v1.5.0",
	PersistentPreRunE: RunRootCmd,
}

func RunRootCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	if err := setUpLogs(l); err != nil {
		return err
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Passes flags to all child/sub commands
	rootCmd.PersistentFlags().StringVarP(&l, "log", "l", "", "log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().StringP("organization", "o", "", "terraform organization or set TFE_ORG")
	rootCmd.PersistentFlags().StringP("token", "t", "", "terraform token or set TFE_TOKEN")
	rootCmd.PersistentFlags().StringP("query", "q", "", "JQ compatible query to parse JSON output")
}

// SetUpLogs sets the log level.
func setUpLogs(level string) error {
	// Read the log level
	//  1. from the CLI first
	//  2. then the ENV vars
	//  3. then use the default value.
	if level == "" {
		level = os.Getenv("TFE_LOG_LEVEL")
		if level == "" {
			level = log.WarnLevel.String()
		}
	}

	// Parse the log level.
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}

	// Set the log level.
	log.SetLevel(lvl)
	return nil
}

func check(err error) {
	if err != nil {
		log.Fatalf("Unable to perform operation: %v\n", err)
	}
}
