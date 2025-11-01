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

package engine

import (
	"fmt"
	"sync"

	"github.com/Purpose-Dev/flowcraft/internal/runner"
)

func Run(graph *Graph, logger *runner.Logger) error {
	levels, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort graph: %w", err)
	}

	totalJobs := len(graph.Nodes)
	logger.Info(fmt.Sprintf("Starting pipeline... %d job(s) to run in %d level(s).", totalJobs, len(levels)))

	for i, level := range levels {
		levelNum := i + 1
		logger.StartGroup(fmt.Sprintf("Level %d/%d (Executing %d job(s) in parallel)", levelNum, len(levels), len(level)))

		var wg sync.WaitGroup
		var firstError error
		var errMutex sync.Mutex

		wg.Add(len(level))

		for _, node := range level {
			go func(n *Node) {
				defer wg.Done()

				err := executeJob(n.Name, n.Job, logger)
				if err != nil {
					errMutex.Lock()
					if firstError == nil {
						firstError = err
					}
					errMutex.Unlock()
				}
			}(node)
		}

		wg.Wait()

		if firstError != nil {
			logger.EndGroup()
			logger.Error(fmt.Sprintf("Failed on level %d: %v", levelNum, firstError))
			return fmt.Errorf("pipeline failed at level %d: %w", levelNum, firstError)
		}

		logger.Success(fmt.Sprintf("Level %d/%d completed successfully.", levelNum, len(levels)))
		logger.EndGroup()
	}

	logger.Success("Pipeline finished successfully. All jobs completed.")
	return nil
}
