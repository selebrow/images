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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dir string
	srv *httptest.Server
)

func init() {
	dir, _ = os.MkdirTemp("", "fileserver")
	srv = httptest.NewServer(mux(dir))
}

func withUrl(path string) string {
	return srv.URL + path
}

func TestDownloadAndRemoveFile(t *testing.T) {
	tempFile, _ := os.CreateTemp(dir, "fileserver")
	_ = os.WriteFile(tempFile.Name(), []byte("test-data"), 0644)
	tempFileName := filepath.Base(tempFile.Name())
	resp, err := http.Get(withUrl("/" + tempFileName))
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	_, err = os.Stat(tempFile.Name())
	assert.Nil(t, err)

	rsp, err := http.Get(withUrl("/?json"))
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	var files []string
	assert.Nil(t, json.NewDecoder(rsp.Body).Decode(&files))
	assert.Equal(t, files, []string{tempFileName})

	req, _ := http.NewRequest(http.MethodDelete, withUrl("/"+tempFileName), nil)
	resp, err = http.DefaultClient.Do(req)

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	_, err = os.Stat(tempFile.Name())
	assert.NotNil(t, err)
}

func TestRemoveMissingFile(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, withUrl("/missing-file"), nil)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}
