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

type Settings struct {
	Parallelism int `toml:"parallelism"`
}

type Secret struct {
	Provider string `toml:"provider"`
	Key      string `toml:"key"`
}

type Config struct {
	Settings Settings          `toml:"settings"`
	Secrets  map[string]Secret `toml:"secrets"`
	Jobs     map[string]Job    `toml:"jobs"`
	Env      map[string]string `toml:"env"`
}

type Job struct {
	Env       map[string]string `toml:"env"`
	Steps     []Step            `toml:"steps"`
	Parallel  []Step            `toml:"parallel"`
	DependsOn []string          `toml:"depends_on"`
	Secrets   []string          `toml:"secrets"`
	Retry     int               `toml:"retry"`
	Inputs    []string          `toml:"inputs"`
	Outputs   []string          `toml:"outputs"`
}

type Step struct {
	Name string `toml:"name"`
	Cmd  string `toml:"cmd"`
	Dir  string `toml:"dir"`
}
