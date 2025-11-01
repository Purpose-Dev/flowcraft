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
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Purpose-Dev/flowcraft/internal/config"
)

func newTestConfig(jobs map[string]config.Job) *config.Config {
	return &config.Config{Jobs: jobs}
}

func TestBuildDag_InvalidDependency(t *testing.T) {
	cfg := newTestConfig(map[string]config.Job{
		"A": {DependsOn: []string{"Z"}},
	})

	_, err := BuildDag(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid dependency, got nil")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error to mention 'does not exist', got: %v", err)
	}
}

func TestBuildDag_SimpleCycle(t *testing.T) {
	cfg := newTestConfig(map[string]config.Job{
		"A": {DependsOn: []string{"B"}},
		"B": {DependsOn: []string{"A"}},
	})

	_, err := BuildDag(cfg) //
	if err == nil {
		t.Fatal("Expected error for cycle, got nil")
	}

	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("Expected error to mention 'circular dependency', got: %v", err)
	}
}

func TestBuildDag_ComplexCycle(t *testing.T) {
	cfg := newTestConfig(map[string]config.Job{
		"A": {DependsOn: []string{"B"}},
		"B": {DependsOn: []string{"C"}},
		"C": {DependsOn: []string{"A"}},
	})

	_, err := BuildDag(cfg) //
	if err == nil {
		t.Fatal("Expected error for cycle, got nil")
	}

	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("Expected error to mention 'circular dependency', got: %v", err)
	}
}

func TestTopologicalSort(t *testing.T) {
	// This tests the exact graph from our test-project
	cfg := newTestConfig(map[string]config.Job{
		"setup":        {DependsOn: []string{}},
		"build-api":    {DependsOn: []string{"setup"}},
		"build-webapp": {DependsOn: []string{"setup"}},
		"test-api":     {DependsOn: []string{"build-api"}},
		"deploy":       {DependsOn: []string{"test-api", "build-webapp"}},
	})

	graph, err := BuildDag(cfg) //
	if err != nil {
		t.Fatalf("Failed to build valid DAG: %v", err)
	}

	levels, err := graph.TopologicalSort() //
	if err != nil {
		t.Fatalf("Failed to sort valid DAG: %v", err)
	}

	getLevelNames := func(level []*Node) []string {
		names := make([]string, len(level))
		for i, node := range level {
			names[i] = node.Name
		}
		sort.Strings(names)
		return names
	}

	if len(levels) != 4 {
		t.Fatalf("Expected 4 levels, got %d", len(levels))
	}

	expectedLevel1 := []string{"setup"}
	if !reflect.DeepEqual(getLevelNames(levels[0]), expectedLevel1) {
		t.Errorf("Expected level 1 to be %v, got %v", expectedLevel1, getLevelNames(levels[0]))
	}

	expectedLevel2 := []string{"build-api", "build-webapp"}
	if !reflect.DeepEqual(getLevelNames(levels[1]), expectedLevel2) {
		t.Errorf("Expected level 2 to be %v, got %v", expectedLevel2, getLevelNames(levels[1]))
	}

	expectedLevel3 := []string{"test-api"}
	if !reflect.DeepEqual(getLevelNames(levels[2]), expectedLevel3) {
		t.Errorf("Expected level 3 to be %v, got %v", expectedLevel3, getLevelNames(levels[2]))
	}

	expectedLevel4 := []string{"deploy"}
	if !reflect.DeepEqual(getLevelNames(levels[3]), expectedLevel4) {
		t.Errorf("Expected level 4 to be %v, got %v", expectedLevel4, getLevelNames(levels[3]))
	}
}
