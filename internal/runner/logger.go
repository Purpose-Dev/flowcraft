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
	"os"
)

const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorReset  = "\033[0m"
)

type Logger struct {
	isCI bool
}

func NewLogger() *Logger {
	isCI := os.Getenv("GITHUB_ACTIONS") == "true"
	return &Logger{isCI: isCI}
}

func (l *Logger) Info(msg string) {
	if l.isCI {
		fmt.Println(msg)
	} else {
		fmt.Printf("%s[INFO] %s%s\n", ColorCyan, msg, ColorReset)
	}
}

func (l *Logger) Error(msg string) {
	if l.isCI {
		fmt.Printf("::error::%s\n", msg)
	} else {
		fmt.Printf("%s[ERROR] %s%s\n", ColorRed, msg, ColorReset)
	}
}

func (l *Logger) Success(msg string) {
	if l.isCI {
		fmt.Printf("::success::%s\n", msg)
	} else {
		fmt.Printf("%s[SUCCESS] %s%s\n", ColorGreen, msg, ColorReset)
	}
}

func (l *Logger) StartGroup(title string) {
	if l.isCI {
		fmt.Printf("::group::%s\n", title)
	} else {
		fmt.Printf("\n%s▶ %s%s\n", ColorYellow, title, ColorReset)
		fmt.Println("------------------------------------------------")
	}
}

func (l *Logger) EndGroup() {
	if l.isCI {
		fmt.Println("::end_group::")
	} else {
		fmt.Println("------------------------------------------------")
	}
}
