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

	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/Purpose-Dev/flowcraft/internal/runner"
)

// executeJob runs all steps for a single job.
// It acts as a "micro-orchestrator" for a job.
func executeJob(jobName string, job config.Job, logger *runner.Logger) error {
	logger.StartGroup(fmt.Sprintf("Job: %s", jobName))
	defer logger.EndGroup()

	if len(job.Steps) > 0 {
		logger.Info(fmt.Sprintf("Starting %d sequential steps for '%s'", len(job.Steps), jobName))
		for _, step := range job.Steps {
			if err := runner.Execute(step, logger); err != nil {
				return fmt.Errorf("sequential step '%s' in job '%s' failed: %w", step.Name, jobName, err)
			}
		}
		logger.Success(fmt.Sprintf("All %d sequential steps for job '%s' completed.", len(job.Steps), jobName))
	}

	if len(job.Parallel) > 0 {
		logger.Info(fmt.Sprintf("Starting %d parallel steps for job '%s'", len(job.Parallel), jobName))

		var wg sync.WaitGroup
		var firstError error
		var errMutex sync.Mutex

		wg.Add(len(job.Parallel))

		for _, step := range job.Parallel {
			go func(s config.Step) {
				defer wg.Done()

				err := runner.Execute(s, logger)
				if err != nil {
					errMutex.Lock()
					if firstError == nil {
						firstError = fmt.Errorf("parallel step '%s' in job '%s' failed: %w", s.Name, jobName, err)
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

	logger.Success(fmt.Sprintf("Job '%s' finished successfully.", jobName))
	return nil
}
