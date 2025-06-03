// Copyright 2017+ Aerokube Software OÜ

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	dir, err := downloadsDir()
	if err != nil {
		log.Fatal(err)
	}

	if err := http.ListenAndServe(":8080", mux(dir)); err != nil {
		log.Fatal(err)
	}
}

func downloadsDir() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home/selenium"
	}

	dir := filepath.Join(homeDir, "Downloads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads dir: %v", err)
	}

	return dir, nil
}

const jsonParam = "json"

func mux(dir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteFileIfExists(w, r, dir)
			return
		}

		if _, ok := r.URL.Query()[jsonParam]; ok {
			listFilesAsJson(w, dir)
			return
		}

		http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
	})

	return mux
}

func listFilesAsJson(w http.ResponseWriter, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, info)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	ret := []string{}
	for _, f := range files {
		ret = append(ret, f.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ret)
}

func deleteFileIfExists(w http.ResponseWriter, r *http.Request, dir string) {
	fileName := strings.TrimPrefix(r.URL.Path, "/")
	filePath := filepath.Join(dir, fileName)
	if _, err := os.Stat(filePath); err != nil {
		http.Error(w, fmt.Sprintf("Unknown file %s", fileName), http.StatusNotFound)
		return
	}

	if err := os.Remove(filePath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete file %s: %v", fileName, err), http.StatusInternalServerError)
		return
	}
}
