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
	"os"
	"reflect"
	"testing"

	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/Purpose-Dev/flowcraft/internal/runner"
)

func TestMergeEnvs(t *testing.T) {
	global := map[string]string{
		"GLOBAL_VAR": "global_value",
		"OVERRIDE":   "from_global",
	}
	job := map[string]string{
		"JOB_VAR":  "job_value",
		"OVERRIDE": "from_job",
	}

	expected := map[string]string{
		"GLOBAL_VAR": "global_value",
		"JOB_VAR":    "job_value",
		"OVERRIDE":   "from_job",
	}

	merged := mergeEnvs(global, job)
	if !reflect.DeepEqual(merged, expected) {
		t.Errorf("mergeEnvs() = %v, want %v", merged, expected)
	}
}

func TestResolveSecrets(t *testing.T) {
	os.Setenv("MY_TEST_SECRET_KEY", "secret-value-123")
	os.Setenv("ANOTHER_KEY", "hello")
	defer os.Unsetenv("MY_TEST_SECRET_KEY")
	defer os.Unsetenv("ANOTHER_KEY")

	cfg := &config.Config{
		Secrets: map[string]config.Secret{
			"db_pass": {
				Provider: "env",
				Key:      "MY_TEST_SECRET_KEY",
			},
			"api_key": {
				Provider: "env",
				Key:      "ANOTHER_KEY",
			},
			"missing_secret": {
				Provider: "env",
				Key:      "I_DONT_EXIST",
			},
		},
	}

	logger := runner.NewLogger()

	resolved, values, err := resolveSecrets(cfg, logger)
	if err != nil {
		t.Fatalf("resolveSecrets()_test.go returned an unexpected error: %v", err)
	}

	if resolved["db_pass"] != "secret-value-123" {
		t.Errorf("Expected 'db_pass' to be 'secret-value-123', got '%s'", resolved["db_pass"])
	}
	if resolved["api_key"] != "hello" {
		t.Errorf("Expected 'api_key' to be 'hello', got '%s'", resolved["api_key"])
	}
	if resolved["missing_secret"] != "" {
		t.Errorf("Expected 'missing_secret' to be '', got '%s'", resolved["missing_secret"])
	}

	expectedValues := []string{"secret-value-123", "hello"}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected secret values %v, got %v", expectedValues, values)
	}
}
