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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/stretchr/testify/assert"
)

var (
	srv         *httptest.Server
	devtoolsSrv *httptest.Server
)

func init() {
	srv = httptest.NewServer(root())
	listen = srv.Listener.Addr().String()
	devtoolsSrv = httptest.NewServer(mockDevtoolsMux())
	defaultDevtoolsHost = devtoolsSrv.Listener.Addr().String()
}

func mockDevtoolsMux() http.Handler {
	mux := http.NewServeMux()
	version := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(fmt.Appendf(nil, `{
			"Browser": "Chrome/72.0.3601.0",
			"Protocol-Version": "1.3",
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3601.0 Safari/537.36",
			"V8-Version": "7.2.233",
			"WebKit-Version": "537.36 (@cfede9db1d154de0468cb0538479f34c0755a0f4)",
			"webSocketDebuggerUrl": "ws://%s/devtools/browser/b0b8a4fb-bb17-4359-9533-a8d9f3908bd8"
		}`, defaultDevtoolsHost))
	}

	mux.HandleFunc("/json/version", version)

	listTargets := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(fmt.Appendf(nil, `[ {
  "description": "",
  "devtoolsFrontendUrl": "/devtools/inspector.html?ws=localhost:9222/devtools/page/one",
  "id": "one",
  "title": "Aerokube",
  "type": "page",
  "url": "https://www.aerokube.com/",
  "webSocketDebuggerUrl": "ws://%s/devtools/page/one"
}, {
  "description": "",
  "devtoolsFrontendUrl": "/devtools/inspector.html?ws=localhost:9222/devtools/page/two",
  "id": "two",
  "title": "Aerokube Selenoid",
  "type": "page",
  "url": "https://selenoid.aerokube.com/",
  "webSocketDebuggerUrl": "ws://%s/devtools/page/two"
} ]`, defaultDevtoolsHost, defaultDevtoolsHost))
	}

	mux.HandleFunc("/json/list", listTargets)
	mux.HandleFunc("/json", listTargets)

	mux.HandleFunc("/json/protocol", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("{}"))
	})

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	devtoolsEcho := func(w http.ResponseWriter, r *http.Request) {
		//Echo request ID but omit Method
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			var req rpcc.Request
			if err := json.Unmarshal(message, &req); err != nil {
				panic(err)
			}

			resp := rpcc.Response{
				ID: req.ID,
			}

			if req.Method == "Target.getTargets" {
				result := target.GetTargetsReply{
					TargetInfos: []target.Info{
						{TargetID: "one", Type: "page"},
						{TargetID: "two", Type: "page"},
					},
				}
				resBytes, _ := json.Marshal(result)
				resp.Result = json.RawMessage(resBytes)
			}

			output, err := json.Marshal(resp)
			if err != nil {
				panic(err)
			}

			if err := c.WriteMessage(mt, output); err != nil {
				break
			}
		}
	}

	mux.HandleFunc("/devtools/browser/", devtoolsEcho)
	mux.HandleFunc("/devtools/page/", devtoolsEcho)
	return mux
}

func TestProtocol(t *testing.T) {
	u := fmt.Sprintf("http://%s/json/protocol", srv.Listener.Addr().String())
	resp, err := http.Get(u)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestDevtools(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	browserWs := fmt.Sprintf("ws://%s/", srv.Listener.Addr().String())
	browserConn, err := rpcc.DialContext(ctx, browserWs)
	assert.Nil(t, err)
	defer browserConn.Close()

	browserClient := cdp.NewClient(browserConn)
	targets, err := browserClient.Target.GetTargets(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(targets.TargetInfos), 2)

	defaultPageWs := fmt.Sprintf("ws://%s/page", srv.Listener.Addr().String())
	defaultPageConn, err := rpcc.DialContext(ctx, defaultPageWs)
	assert.Nil(t, err)
	defer defaultPageConn.Close()

	defaultPageClient := cdp.NewClient(defaultPageConn)
	err = defaultPageClient.Page.Enable(ctx)
	assert.Nil(t, err)

	selectedPageWs := fmt.Sprintf("ws://%s/page/%s", srv.Listener.Addr().String(), targets.TargetInfos[1].TargetID)
	selectedPageConn, err := rpcc.DialContext(ctx, selectedPageWs)
	assert.Nil(t, err)
	defer selectedPageConn.Close()

	selectedPageClient := cdp.NewClient(selectedPageConn)
	err = selectedPageClient.Page.Enable(ctx)
	assert.Nil(t, err)
}

func TestDetectDevtoolsHost(t *testing.T) {
	name, _ := os.MkdirTemp("", "devtools")
	defer os.RemoveAll(name)
	profilePath := filepath.Join(name, ".org.chromium.Chromium.deadbee")
	_ = os.MkdirAll(profilePath, os.ModePerm)
	portFile := filepath.Join(profilePath, "DevToolsActivePort")
	_ = os.WriteFile(portFile, []byte("12345\n/devtools/browser/6f37c7fe-a0a6-4346-a6e2-80c5da0687f0"), 0644)

	assert.Equal(t, detectDevtoolsHost(name), "127.0.0.1:12345")
}

func TestDetectDevtoolsHostProfileDir(t *testing.T) {
	name, _ := os.MkdirTemp("", "devtools")
	defer os.RemoveAll(name)
	profilePath := filepath.Join(name, "can_be_anything")
	_ = os.MkdirAll(profilePath, os.ModePerm)
	portFile := filepath.Join(profilePath, "DevToolsActivePort")
	_ = os.WriteFile(portFile, []byte("12345\n/devtools/browser/6f37c7fe-a0a6-4346-a6e2-80c5da0687f0"), 0644)

	t.Setenv("BROWSER_PROFILE_DIR", profilePath)
	assert.Equal(t, detectDevtoolsHost(name), "127.0.0.1:12345")
}
