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

package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("non-existent-file.toml")
	if err == nil {
		t.Fatal("Expected an error for a non-existent file, but got nil.")
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected error to contain 'no such file or directory', got: %v", err)
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "bad_toml_*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte("[jobs.myjob\\n...this is not valid toml")); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected an error for invalid TOML, but got nil")
	}

	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("Expected error to contain 'parsing', got: %v", err)
	}
}
