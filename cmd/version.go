/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is the version of the CLI, to be overwritten by make/goreleaser via ldflags.
var version string

// getVersion determines the most accurate version string available.
// It prioritizes the version injected at build time via ldflags, then falls back to
// the version information embedded by `go install`, and finally defaults to "(devel)".
func getVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the snip CLI version information.",
	Long: `The version command prints the application's version. The version is determined
at compile time: official releases show a tagged version (e.g. v1.0.2), while local
development builds show "(devel)".`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("snip CLI Version: %s\n", getVersion())
	},
}
