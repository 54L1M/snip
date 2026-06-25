/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

//go:embed .snip.example
var exampleConfig embed.FS

// setupCmd represents the setup command.
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create the default configuration file.",
	Long: `Creates the default settings file at ~/.config/snip/.snip.

snip works without it (every setting has a default), but this gives you a
documented file to tweak the editor and confirmation behavior.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := store.ConfigDir()
		if err != nil {
			return err
		}
		configFile := store.ConfigFile()

		if _, err := os.Stat(configFile); err == nil {
			fmt.Println(ui.WarningStyle.Render("Configuration file already exists at " + configFile + "\nSkipping creation to avoid overwriting your settings."))
			return nil
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		data, err := exampleConfig.ReadFile(".snip.example")
		if err != nil {
			return fmt.Errorf("failed to read embedded config file: %w", err)
		}

		if err := os.WriteFile(configFile, data, 0o644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Println("✅", ui.SuccessStyle.Render("Created config file at: "+configFile))
		return nil
	},
}
