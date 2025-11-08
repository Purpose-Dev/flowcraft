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
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/Purpose-Dev/flowcraft/internal/runner"
)

func Run(ctx context.Context, cfg *config.Config, graph *Graph, logger *runner.Logger) error {
	levels, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort graph: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	totalJobs := len(graph.Nodes)
	logger.Info(fmt.Sprintf("Starting pipeline... %d job(s) to run in %d level(s).", totalJobs, len(levels)))

	numWorkers := cfg.Settings.Parallelism
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	logger.Info(fmt.Sprintf("Concurrency limit set to %d worker(s).", numWorkers))

	resolvedSecrets, secretValues, err := resolveSecrets(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to resolve secrets: %w", err)
	}

	logger.SetSecretsToMask(secretValues)

	for i, level := range levels {
		levelNum := i + 1
		logger.StartGroup(fmt.Sprintf("Level %d/%d (Executing %d job(s) in parallel)", levelNum, len(levels), len(level)))

		levelCtx, cancel := context.WithCancel(ctx)

		var wg sync.WaitGroup
		var firstError error
		var errMutex sync.Mutex

		jobQueue := make(chan *Node, len(level))

		wg.Add(len(level))

		for w := 0; w < numWorkers; w++ {
			go func() {
				for node := range jobQueue {
					jobEnvs := mergeEnvs(cfg.Env, node.Job.Env)
					for _, secretName := range node.Job.Secrets {
						if val, ok := resolvedSecrets[secretName]; ok {
							jobEnvs[secretName] = val
						}
					}

					err := executeJob(levelCtx, node.Name, node.Job, jobEnvs, logger)
					if err != nil {
						errMutex.Lock()
						if firstError == nil {
							firstError = err
							cancel()
						}
						errMutex.Unlock()
					}
					wg.Done()
				}
			}()
		}

		for _, node := range level {
			jobQueue <- node
		}

		close(jobQueue)

		wg.Wait()
		cancel()

		if firstError != nil {
			logger.EndGroup()
			if firstError == context.Canceled {
				logger.Error(fmt.Sprintf("Level %d failed and was cancelled.", levelNum))
				return fmt.Errorf("pipeline failed at level %d", levelNum)
			}
			logger.Error(fmt.Sprintf("Failed on level %d: %v", levelNum, firstError))
			return fmt.Errorf("pipeline failed at level %d: %w", levelNum, firstError)
		}

		if err := ctx.Err(); err != nil {
			logger.EndGroup()
			return err
		}

		logger.Success(fmt.Sprintf("Level %d/%d completed successfully.", levelNum, len(levels)))
		logger.EndGroup()
	}

	logger.Success("Pipeline finished successfully. All jobs completed.")
	return nil
}

func resolveSecrets(cfg *config.Config, logger *runner.Logger) (map[string]string, []string, error) {
	resolved := make(map[string]string)
	var valuesToMask []string

	for logicalName, secret := range cfg.Secrets {
		if secret.Provider != "env" {
			logger.Error(fmt.Sprintf("Secret '%s': provider '%s' is not supported (only 'env' is supported in v0.3.0)", logicalName, secret.Provider))
			continue
		}
		if secret.Key == "" {
			return nil, nil, fmt.Errorf("secret '%s': 'key' in config cannot be empty", logicalName)
		}
		val, exists := os.LookupEnv(secret.Key)
		if !exists {
			logger.Error(fmt.Sprintf("Secret '%s': environment variable '%s' is not set", logicalName, secret.Key))
		}

		resolved[logicalName] = val
		if val != "" {
			valuesToMask = append(valuesToMask, val)
		}
	}

	return resolved, valuesToMask, nil
}

func mergeEnvs(globalEnv, jobEnv map[string]string) map[string]string {
	merged := make(map[string]string)

	for k, v := range globalEnv {
		merged[k] = v
	}

	for k, v := range jobEnv {
		merged[k] = v
	}

	return merged
}
