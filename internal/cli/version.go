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
	"time"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version, commit, and build date of flowcraft",
	Long:  "Prints, the version, commit, and build date of the flowcraft binary.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flowcraft v%s\n", version)
		fmt.Printf("Commit %s\n", commit)

		builtDate := date
		t, err := time.Parse(time.RFC3339, builtDate)
		if err == nil {
			builtDate = t.Format("02 Jan 2006 at 15:04 MST")
		}

		fmt.Printf("Built: %s\n", builtDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
