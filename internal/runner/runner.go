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

package runner

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"

	"github.com/Purpose-Dev/flowcraft/internal/config"
)

func Execute(step config.Step, logger *Logger) error {
	logger.StartGroup(fmt.Sprintf("Step: %s", step.Name))
	defer logger.EndGroup()

	logger.Info(fmt.Sprintf("Executing command: %s", step.Cmd))
	cmd := exec.Command("bash", "-c", step.Cmd)
	if step.Dir != "" {
		cmd.Dir = step.Dir
		logger.Info(fmt.Sprintf("Working directory: %s", step.Dir))
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe for step '%s': %w", step.Name, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe for step '%s': %w", step.Name, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start step '%s': %w", step.Name, err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			logger.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logger.Error(fmt.Sprintf("Error scanning stdout for step '%s': %v\n", step.Name, err))
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logger.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			logger.Error(fmt.Sprintf("Error scanning stderr for step '%s': %v\n", step.Name, err))
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		wrappedError := fmt.Errorf("step '%s' failed: %w", step.Name, err)
		logger.Error(wrappedError.Error())
		return wrappedError
	}

	logger.Success(fmt.Sprintf("Step '%s' completed successfully", step.Name))
	return nil
}
