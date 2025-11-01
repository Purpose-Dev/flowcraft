/*
 * Copyright 2025 Riyane El Qoqui
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"fmt"
	"log"

	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/Purpose-Dev/flowcraft/internal/engine"
	"github.com/Purpose-Dev/flowcraft/internal/runner"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the flow.toml configuration file",
	Long: `Parses the configuration file and builds the dependency graph (DAG)
to check for syntax errors and circular dependencies.
This command does not execute any jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runner.NewLogger()

		filePath, _ := cmd.Flags().GetString("file")
		logger.Info(fmt.Sprintf("Validating configuration from: %s", filePath))

		cfg, err := config.LoadConfig(filePath)
		if err != nil {
			logger.Error(fmt.Sprintf("Configuration validation failed (parsing error): %v", err))
			log.Fatalf("Validation failed: %v", err)
		}

		_, err = engine.BuildDag(cfg)
		if err != nil {
			logger.Error(fmt.Sprintf("DAG validation failed (e.g., circular dependency): %v", err))
			log.Fatalf("Validation failed: %v", err)
		}

		logger.Success("Validation OK. Configuration is valid and no cycles were found.")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringP("file", "f", "flow.toml", "Path to the flow.toml configuration file")
}
