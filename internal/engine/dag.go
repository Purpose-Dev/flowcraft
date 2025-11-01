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

	"github.com/Purpose-Dev/flowcraft/internal/config"
)

type Node struct {
	Name         string
	Job          config.Job
	Dependencies []*Node
	Dependents   []*Node
}

type Graph struct {
	Nodes map[string]*Node
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
	}
}

// BuildDag creates a dependency graph from the configuration.
// It also detects circular dependencies.
func BuildDag(cfg *config.Config) (*Graph, error) {
	graph := NewGraph()

	for jobName, job := range cfg.Jobs {
		graph.Nodes[jobName] = &Node{
			Name:         jobName,
			Job:          job,
			Dependencies: []*Node{}, // Init empty
			Dependents:   []*Node{}, // Init empty
		}
	}

	for jobName, node := range graph.Nodes {
		for _, depName := range node.Job.DependsOn {
			depNode, exists := graph.Nodes[depName]
			if !exists {
				return nil, fmt.Errorf(
					"job '%s' has an invalid dependency: '%s' does not exist",
					jobName, depName,
				)
			}
			node.Dependencies = append(node.Dependencies, depNode)
			depNode.Dependents = append(depNode.Dependents, node)
		}
	}

	if err := graph.detectCycles(); err != nil {
		return nil, fmt.Errorf("invalid pipeline: %w", err)
	}

	return graph, nil
}

// detectCycles performs a Depth-First Search (DFS) to find cycles.
func (g *Graph) detectCycles() error {
	visiting := make(map[string]bool)
	visited := make(map[string]bool)

	var path []string
	var visit func(node *Node) error

	visit = func(node *Node) error {
		path = append(path, node.Name)

		if visiting[node.Name] {
			cyclePath := fmt.Sprintf("%s -> %s", node.Name, path[0])
			for i := 0; i < len(path); i++ {
				cyclePath = fmt.Sprintf("%s -> %s", path[i], cyclePath)
			}
			for i, pNode := range path {
				if pNode == node.Name {
					cyclePath = fmt.Sprintf("%s", path[i])
					for j := i + 1; j < len(path); j++ {
						cyclePath = fmt.Sprintf("%s -> %s", cyclePath, path[j])
					}
					break
				}
			}

			return fmt.Errorf("circular dependency detected: %s", cyclePath)
		}

		if visited[node.Name] {
			path = path[:len(path)-1]
			return nil
		}

		visiting[node.Name] = true

		for _, dep := range node.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[node.Name] = false
		visited[node.Name] = true
		path = path[:len(path)-1]

		return nil
	}

	for jobName, node := range g.Nodes {
		if !visited[jobName] {
			if err := visit(node); err != nil {
				return err
			}
		}
	}

	return nil
}

// TopologicalSort performs a topological sort on the graph using Kahn's algorithm.
// It returns the nodes in "levels" of execution.
// Example: [ [A], [B, C], [D] ]
// This means A runs first, then B and C run parallelly, then D runs.
func (g *Graph) TopologicalSort() ([][]*Node, error) {
	inDegree := make(map[string]int)
	for name, node := range g.Nodes {
		inDegree[name] = len(node.Dependencies)
	}

	var queue []*Node
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, g.Nodes[name])
		}
	}

	var levels [][]*Node

	for len(queue) > 0 {
		currentLevel := make([]*Node, len(queue))
		copy(currentLevel, queue)
		levels = append(levels, currentLevel)

		var nextQueue []*Node

		for _, node := range queue {
			for _, dependent := range node.Dependents {
				depName := dependent.Name
				inDegree[depName]--

				if inDegree[depName] == 0 {
					nextQueue = append(nextQueue, dependent)
				}
			}
		}

		queue = nextQueue
	}

	visitedCount := 0
	for _, level := range levels {
		visitedCount += len(level)
	}

	if visitedCount != len(g.Nodes) {
		return nil, fmt.Errorf(
			"graph has a cycle (topological sort failed to visit all nodes, %d vs %d)",
			visitedCount, len(g.Nodes),
		)
	}

	return levels, nil
}
