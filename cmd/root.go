/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/54L1M/snip/internal/ui"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "snip [command]",
	Short: "Save long, parameterized commands as snippets and run them with ease.",
	Long: `snip stores long shell commands as named templates with {{placeholders}},
then resolves and runs them with the changing parts supplied as key=value
arguments or filled in interactively. The fully-resolved command is always
shown for confirmation before it runs, and the last value used for each
variable is remembered as the next default.`,
	Version: getVersion(),
	Run:     func(cmd *cobra.Command, args []string) { ui.DisplayWelcomeScreen() },
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	ui.ApplyCustomHelpTemplate(rootCmd)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/snip/.snip)")
}

// initConfig reads in the config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(filepath.Join(home, ".config", "snip"))
		viper.SetConfigType("env")
		viper.SetConfigName(".snip")
	}

	viper.AutomaticEnv()

	// A missing config file is fine — `snip setup` creates one, and all
	// settings have sensible defaults.
	_ = viper.ReadInConfig()
}
