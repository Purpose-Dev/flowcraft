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
	"strings"
	"sync"
)

const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorReset  = "\033[0m"
)

type Logger struct {
	out           io.Writer
	mu            sync.Mutex
	isCI          bool
	secretsToMask []string
	err           error
}

func NewLogger() *Logger {
	isCI := os.Getenv("GITHUB_ACTIONS") == "true"
	return &Logger{isCI: isCI, out: os.Stdout}
}

func (l *Logger) GetWriter() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

func (l *Logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func (l *Logger) SetSecretsToMask(secrets []string) {
	for _, s := range secrets {
		if s != "" {
			l.secretsToMask = append(l.secretsToMask, s)
		}
	}
}

func (l *Logger) Err() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.err
}

func (l *Logger) scrub(msg string) string {
	if len(l.secretsToMask) == 0 {
		return msg
	}
	for _, secret := range l.secretsToMask {
		msg = strings.ReplaceAll(msg, secret, "[SECRET]")
	}
	return msg
}

func (l *Logger) write(data []byte) {
	if l.err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := l.out.Write(data)
	if err != nil && l.err == nil {
		l.err = err
		_, _ = fmt.Fprintf(os.Stderr, "%s[LOGGER ERROR] Failed to write logs: %v%s\n", ColorRed, err, ColorReset)
	}
}

func (l *Logger) Info(msg string) {
	msg = l.scrub(msg)
	var formattedMsg string
	if l.isCI {
		formattedMsg = msg + "\n"
	} else {
		formattedMsg = fmt.Sprintf("%s[INFO] %s%s\n", ColorCyan, msg, ColorReset)
	}
	l.write([]byte(formattedMsg))
}

func (l *Logger) Warn(msg string) {
	msg = l.scrub(msg)
	var formattedMsg string
	if l.isCI {
		formattedMsg = fmt.Sprintf("::warn::%s\n", msg)
	} else {
		formattedMsg = fmt.Sprintf("%s[WARN] %s%s\n", ColorYellow, msg, ColorReset)
	}
	l.write([]byte(formattedMsg))
}

func (l *Logger) Error(msg string) {
	msg = l.scrub(msg)
	var formattedMsg string
	if l.isCI {
		formattedMsg = fmt.Sprintf("::error::%s\n", msg)
	} else {
		formattedMsg = fmt.Sprintf("%s[ERROR] %s%s\n", ColorRed, msg, ColorReset)
	}
	l.write([]byte(formattedMsg))
}

func (l *Logger) Success(msg string) {
	msg = l.scrub(msg)
	var formattedMsg string
	if l.isCI {
		formattedMsg = fmt.Sprintf("::success::%s\n", msg)
	} else {
		formattedMsg = fmt.Sprintf("%s[SUCCESS] %s%s\n", ColorGreen, msg, ColorReset)
	}
	l.write([]byte(formattedMsg))
}

func (l *Logger) StartGroup(title string) {
	title = l.scrub(title)
	var formattedMsg string
	if l.isCI {
		formattedMsg = fmt.Sprintf("::group::%s\n", title)
	} else {
		formattedMsg = fmt.Sprintf("\n%s▶ %s%s\n%s\n", ColorYellow, title,
			ColorReset, "------------------------------------------------",
		)
	}
	l.write([]byte(formattedMsg))
}

func (l *Logger) EndGroup() {
	var formattedMsg string
	if l.isCI {
		formattedMsg = fmt.Sprintf("::end_group::\n")
	} else {
		formattedMsg = fmt.Sprintf("------------------------------------------------\n")
	}
	l.write([]byte(formattedMsg))
}

func (l *Logger) Replay(logs []byte) {
	l.write(logs)
}
