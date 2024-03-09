package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	local_config2 "github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"os"
)

// TODO: System config + db related, not related to local config

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
	systemConfigCopy := config.DeepCopy()
	if systemConfigCopy == nil {
		return fmt.Errorf("failed to deep copy system config file")
	}
	return runPatch(systemConfigCopy, []func(*local_config2.Config) error{
		saveUpdatedSystemConfig,
	})
}

// Patches list
func saveUpdatedSystemConfig(config *local_config2.Config) error {
	return local_config2.Update(config)
}

// private function
func runPatch(config *local_config2.Config, listedPatches []func(*local_config2.Config) error) error {
	for _, patch := range listedPatches {
		err := patch(config)
		if err != nil {
			return err
		}
	}
	return nil
}
