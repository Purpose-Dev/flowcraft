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

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a flowcraft pipeline from a configuration file",
	Long: `Executes a flowcraft pipeline by reading a flow.toml file,
building the dependency graph (DAG), and executing the jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runner.NewLogger()
		logger.Info("Flowcraft execution started.")

		filePath, _ := cmd.Flags().GetString("file")
		logger.Info(fmt.Sprintf("Loading configuration from: %s", filePath))

		cfg, err := config.LoadConfig(filePath)
		if err != nil {
			logger.Error(fmt.Sprintf("Error loading configuration: %v", err))
			log.Fatalf("Critical error: %v", err)
		}
		logger.Info(fmt.Sprintf("Configuration loaded successfully. Found %d job(s).", len(cfg.Jobs)))

		logger.Info("Building dependency graph (DAG)...")

		graph, err := engine.BuildDag(cfg)
		if err != nil {
			logger.Error(fmt.Sprintf("Error building DAG: %v", err))
			log.Fatalf("Critical error: %v", err)
		}
		logger.Success("DAG built and validated successfully (no cycles found).")

		if err := engine.Run(graph, logger); err != nil {
			logger.Error(fmt.Sprintf("Pipeline execution failed: %v", err))
			log.Fatalf("Critical error: %v", err)
		}

		logger.Success("Flowcraft execution finished successfully.")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("file", "f", "flow.toml", "Path to the flow.toml configuration file")
}
