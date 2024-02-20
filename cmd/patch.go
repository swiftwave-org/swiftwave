package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/system_config"
	"os"
)

var applyPatchesCmd = &cobra.Command{
	Use:   "apply-patches",
	Short: "Apply patches to swiftwave",
	Long:  `Apply patches to swiftwave`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ApplyPatches()
		if err != nil {
			printError("Failed to apply patches")
			printError(err.Error())
			os.Exit(1)
		} else {
			printSuccess("Patches applied successfully")
		}
	},
}

func ApplyPatches() error {
	// deep copy system config
	systemConfigCopy := systemConfig.DeepCopy()
	if systemConfigCopy == nil {
		return fmt.Errorf("failed to deep copy system config file")
	}
	return runPatch(systemConfigCopy, []func(*system_config.Config) error{
		saveUpdatedSystemConfig,
	})
}

// Patches list
func saveUpdatedSystemConfig(config *system_config.Config) error {
	return config.WriteToFile(configFilePath)
}

// private function
func runPatch(config *system_config.Config, listedPatches []func(*system_config.Config) error) error {
	for _, patch := range listedPatches {
		err := patch(config)
		if err != nil {
			return err
		}
	}
	return nil
}
