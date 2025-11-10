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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Purpose-Dev/flowcraft/internal/config"
	"github.com/bmatcuk/doublestar/v4"
)

const (
	logsFileName     = "logs.txt"
	outputsFileName  = "outputs.tar.gz"
	metadataFileName = "metadata.json"
)

type Store struct {
	cacheDir string
}

func NewStore(rootDir string) (*Store, error) {
	cacheDir := filepath.Join(rootDir, ".flowcraft", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	return &Store{cacheDir: cacheDir}, nil
}

func (s *Store) CalculateKey(job config.Job, envVars map[string]string) (string, error) {
	return calculateCacheKey(job, envVars)
}

func (s *Store) Has(key string) bool {
	entryDir := s.getEntryDir(key)

	if _, err := os.Stat(filepath.Join(entryDir, metadataFileName)); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(filepath.Join(entryDir, logsFileName)); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(filepath.Join(entryDir, outputsFileName)); os.IsNotExist(err) {
		return false
	}

	return true
}

func (s *Store) GetLogs(key string) ([]byte, error) {
	logFile := filepath.Join(s.getEntryDir(key), logsFileName)
	return os.ReadFile(logFile)
}

func (s *Store) StoreNewEntry(key string, outputs []string, logs []byte) error {
	entryDir := s.getEntryDir(key)
	if err := os.MkdirAll(entryDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache entry dir: %w", err)
	}

	logFile := filepath.Join(entryDir, logsFileName)
	if err := os.WriteFile(logFile, logs, 0644); err != nil {
		return fmt.Errorf("failed to write cache logs: %w", err)
	}

	tarballPath := filepath.Join(entryDir, outputsFileName)
	if err := createTarball(tarballPath, outputs); err != nil {
		return fmt.Errorf("failed to create cache tarball: %w", err)
	}

	metaFile := filepath.Join(entryDir, metadataFileName)
	metaData := fmt.Sprintf(`{"createdAt": "%s"}`, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(metaFile, []byte(metaData), 0644); err != nil {
		return fmt.Errorf("failed to write cache metadata: %w", err)
	}

	return nil
}

func (s *Store) RestoreOutputs(key string) error {
	tarballPath := filepath.Join(s.getEntryDir(key), outputsFileName)

	file, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func(gzReader *gzip.Reader) {
		err := gzReader.Close()
		if err != nil {
			return
		}
	}(gzReader)

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) getEntryDir(key string) string {
	return filepath.Join(s.cacheDir, key)
}

func createTarball(tarballPath string, patterns []string) error {
	var filesToArchive []string
	for _, pattern := range patterns {
		files, err := doublestar.Glob(os.DirFS("."), pattern)
		if err != nil {
			return err
		}
		filesToArchive = append(filesToArchive, files...)
	}

	tarFile, err := os.Create(tarballPath)
	if err != nil {
		return err
	}
	defer func(tarFile *os.File) {
		err := tarFile.Close()
		if err != nil {
			return
		}
	}(tarFile)

	gzWriter := gzip.NewWriter(tarFile)
	defer func(gzWriter *gzip.Writer) {
		err := gzWriter.Close()
		if err != nil {
			return
		}
	}(gzWriter)

	tarWriter := tar.NewWriter(gzWriter)
	defer func(tarWriter *tar.Writer) {
		err := tarWriter.Close()
		if err != nil {
			return
		}
	}(tarWriter)

	for _, path := range filesToArchive {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		header.Name = path
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tarWriter, file); err != nil {
			if err := file.Close(); err != nil {
				return err
			}
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}
