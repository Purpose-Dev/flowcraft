/*
 * Copyright 2025 Riyane El Qoqui
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/Purpose-Dev/flowcraft/internal/cache"
	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/Purpose-Dev/flowcraft/internal/runner"
)

// executeJob runs all steps for a single job.
// It acts as a "micro-orchestrator" for a job.
func executeJob(
	ctx context.Context,
	jobName string,
	job config.Job,
	envVars map[string]string,
	logger *runner.Logger,
	store *cache.Store,
) error {
	isCacheable := len(job.Inputs) > 0 && len(job.Outputs) > 0
	var cacheKey string
	var err error

	if isCacheable {
		cacheKey, err = store.CalculateKey(job, envVars)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to calculate cache key for %s: %v. Job will run without cache.", jobName, err))
			isCacheable = false
		}

		if isCacheable && store.Has(cacheKey) {
			logger.StartGroup(fmt.Sprintf("Job: %s (Restoring from cache)", jobName))
			defer logger.EndGroup()

			if err := store.RestoreOutputs(cacheKey); err != nil {
				logger.Error(fmt.Sprintf("Cache HIT but restore failed: %v. Re-running job.", err))
			} else {
				logs, _ := store.GetLogs(cacheKey)
				logger.Replay(logs)
				logger.Success(fmt.Sprintf("Job '%s' finished (FROM CACHE)", jobName))
				return nil
			}
		}
	}

	originalWriter := logger.GetWriter()
	logCapture := new(bytes.Buffer)

	logger.SetWriter(io.MultiWriter(originalWriter, logCapture))

	defer func() {
		logger.SetWriter(originalWriter)
	}()

	logger.StartGroup(fmt.Sprintf("Job: %s", jobName))
	runError := runStepsInJob(ctx, jobName, job, envVars, logger)
	logger.EndGroup()

	if isCacheable && runError == nil {
		logger.Info("Saving outputs to cache...")
		err := store.StoreNewEntry(cacheKey, job.Outputs, logCapture.Bytes())
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to save cache: %v", err))
		}
	}

	return runError
}

func runStepsInJob(ctx context.Context, jobName string, job config.Job, envVars map[string]string, logger *runner.Logger) error {
	if len(job.Steps) > 0 {
		logger.Info(fmt.Sprintf("Starting %d sequential steps for '%s'", len(job.Steps), jobName))
		for _, step := range job.Steps {
			if err := runner.Execute(ctx, step, envVars, logger); err != nil {
				return fmt.Errorf("sequential step '%s' in job '%s' failed: %w", step.Name, jobName, err)
			}
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		logger.Success(fmt.Sprintf("All %d sequential steps for job '%s' completed.", len(job.Steps), jobName))
	}

	if len(job.Parallel) > 0 {
		logger.Info(fmt.Sprintf("Starting %d parallel steps for job '%s'", len(job.Parallel), jobName))

		jobCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var wg sync.WaitGroup
		var firstError error
		var errMutex sync.Mutex

		wg.Add(len(job.Parallel))

		for _, step := range job.Parallel {
			go func(s config.Step) {
				defer wg.Done()

				err := runner.Execute(jobCtx, s, envVars, logger)
				if err != nil {
					errMutex.Lock()
					if firstError == nil {
						firstError = fmt.Errorf("parallel step '%s' in job '%s' failed: %w", s.Name, jobName, err)
						cancel()
					}
					errMutex.Unlock()
				}
			}(step)
		}

		wg.Wait()

		if firstError != nil {
			return firstError
		}

		logger.Success(fmt.Sprintf("All %d parallel steps for job '%s' completed.", len(job.Parallel), jobName))
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	logger.Success(fmt.Sprintf("Job '%s' finished successfully.", jobName))
	return nil
}
