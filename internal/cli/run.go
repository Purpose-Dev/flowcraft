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

		filePath, _ := cmd.Flags().GetString("file")
		logger.Info(fmt.Sprintf("Loading configuration from: %s", filePath))

		_, err := config.LoadConfig(filePath)
		if err != nil {
			logger.Error(fmt.Sprintf("Error loading configuration: %v", err))
			log.Fatalf("Critical error: %v", err)
		}

		logger.Info("Configuration loaded successfully.")

		testStep := config.Step{
			Name: "Test Runner (Local)",
			Cmd:  "echo 'Hello from the runner ^_^ !' && sleep 1 && echo 'Test step passed'",
			Dir:  "",
		}

		if err := runner.Execute(testStep, logger); err != nil {
			logger.Error(fmt.Sprintf("Test step failed: %v", err))
		}

		failingStep := config.Step{
			Name: "Test Failing Runner",
			Cmd:  "echo 'This command will fail' && exit 1",
			Dir:  "",
		}

		if err := runner.Execute(failingStep, logger); err != nil {
			logger.Error(fmt.Sprintf("Failing test step finished (as expected)"))
		} else {
			logger.Success("Failing test step finished (UNEXPECTEDLY)")
		}

		/*for jobName := range cfg.Jobs {
			fmt.Printf("    - Found job: %s\n", jobName)
		}*/
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("file", "f", "flow.toml", "Path to the flow.toml configuration file")
}
