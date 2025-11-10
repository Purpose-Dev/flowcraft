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

package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/bmatcuk/doublestar/v4"
)

type stableJobDefinition struct {
	Steps    []config.Step     `json:"steps"`
	Parallel []config.Step     `json:"parallel"`
	Env      map[string]string `json:"env"`
	Inputs   []string          `json:"inputs"`
	Outputs  []string          `json:"outputs"`
	Secrets  []string          `json:"secrets"`
}

func calculateCacheKey(job config.Job, envVars map[string]string) (string, error) {
	inputHash, err := hashInputFiles(job.Inputs)
	if err != nil {
		return "", fmt.Errorf("failed to hash input files: %w", err)
	}

	jobHash, err := hashJobDefinition(job, envVars)
	if err != nil {
		return "", fmt.Errorf("failed to hash job definition: %w", err)
	}

	finalHasher := sha256.New()
	finalHasher.Write([]byte(inputHash))
	finalHasher.Write([]byte(jobHash))

	return hex.EncodeToString(finalHasher.Sum(nil)), nil
}

func hashJobDefinition(job config.Job, envVars map[string]string) (string, error) {
	sort.Strings(job.Secrets)

	stableDef := stableJobDefinition{
		Steps:    job.Steps,
		Parallel: job.Parallel,
		Env:      envVars,
		Inputs:   job.Inputs,
		Outputs:  job.Outputs,
		Secrets:  job.Secrets,
	}

	jsonData, err := json.Marshal(stableDef)
	if err != nil {
		return "", err
	}

	hasher := sha256.New()
	hasher.Write(jsonData)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashInputFiles(inputs []string) (string, error) {
	if len(inputs) == 0 {
		return "", nil
	}

	var allFiles []string
	for _, pattern := range inputs {
		files, err := doublestar.Glob(os.DirFS("."), pattern)
		if err != nil {
			return "", fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}
		allFiles = append(allFiles, files...)
	}

	sort.Strings(allFiles)
	manifest := &strings.Builder{}

	for _, file := range allFiles {
		info, err := os.Stat(file)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			continue
		}

		hash, err := hashFile(file)
		if err != nil {
			return "", err
		}
		_, _ = fmt.Fprintf(manifest, "%s:%s\n", file, hash)
	}

	hasher := sha256.New()
	hasher.Write([]byte(manifest.String()))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
