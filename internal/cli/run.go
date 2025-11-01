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
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a flowcraft pipeline from a configuration file",
	Long: `Executes a flowcraft pipeline by reading a flow.toml file,
building the dependency graph (DAG), and executing the jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")

		fmt.Printf("Loading configuration from: %s\n", filePath)

		cfg, err := config.LoadConfig(filePath)
		if err != nil {
			log.Fatalf("Error loading configuration:\n%v\n", err)
		}

		log.Printf("Configuration loaded successfully. Found %d job(s).\n", len(cfg.Jobs))
		for jobName := range cfg.Jobs {
			fmt.Printf("    - Found job: %s\n", jobName)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("file", "f", "flow.toml", "Path to the flow.toml configuration file")
}
