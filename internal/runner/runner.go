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
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/Purpose-Dev/flowcraft/internal/config"
)

func Execute(step config.Step) error {
	fmt.Printf("--- Executing Step: %s ---\n", step.Name)
	fmt.Printf("CMD: %s\n", step.Cmd)

	cmd := exec.Command("bash", "-c", step.Cmd)
	if step.Dir != "" {
		cmd.Dir = step.Dir
		fmt.Printf("DIR: %s\n", step.Dir)
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
		if _, err := io.Copy(os.Stdout, stdoutPipe); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error copying stdout for step '%s': %v\n", step.Name, err)
		}
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(os.Stderr, stderrPipe); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error copying stderr for step '%s': %v\n", step.Name, err)
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("step '%s' failed: %w", step.Name, err)
	}

	fmt.Printf("--- Step Succeeded: %s ---\n", step.Name)
	return nil
}
