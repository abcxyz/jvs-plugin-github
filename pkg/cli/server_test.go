// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abcxyz/jvs-plugin-github/pkg/plugin/keyutil"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
)

const (
	testGitHubPluginDisplayName = "test DisplayName"
	testGitHubPluginHint        = "test Hint"
)

func TestServerCommand(t *testing.T) {
	t.Parallel()

	testRSAPrivateKeyString, _ := keyutil.TestGenerateRSAPrivateKey(t)

	ctx := logging.WithLogger(t.Context(), logging.TestLogger(t))

	fakeGitHub := func() *httptest.Server {
		mux := http.NewServeMux()
		mux.Handle("GET /app/installations/123", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"access_tokens_url": "http://%s/app/installations/123/access_tokens"}`, r.Host)
		}))
		mux.Handle("POST /app/installations/123/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			fmt.Fprintf(w, `{"token": "this-is-the-token-from-github"}`)
		}))
		return httptest.NewServer(mux)
	}()
	t.Cleanup(func() {
		fakeGitHub.Close()
	})

	cases := []struct {
		name   string
		args   []string
		env    map[string]string
		expErr string
	}{
		{
			name: "success",
			env: map[string]string{
				"GITHUB_APP_ID":              "my-app",
				"GITHUB_APP_INSTALLATION_ID": "123",
				"GITHUB_APP_PRIVATE_KEY_PEM": testRSAPrivateKeyString,
				"GITHUB_PLUGIN_DISPLAY_NAME": testGitHubPluginDisplayName,
				"GITHUB_PLUGIN_HINT":         testGitHubPluginHint,
				"GITHUB_API_BASE_URL":        fakeGitHub.URL,
			},
		},
		{
			name:   "unexpected args",
			args:   []string{"foo"},
			expErr: `unexpected arguments: ["foo"]`,
		},
		{
			name: "invalid_config_missing_github_app_id",
			env: map[string]string{
				"GITHUB_APP_INSTALLATION_ID": "123",
				"GITHUB_APP_PRIVATE_KEY_PEM": testRSAPrivateKeyString,
				"GITHUB_PLUGIN_DISPLAY_NAME": testGitHubPluginDisplayName,
				"GITHUB_PLUGIN_HINT":         testGitHubPluginHint,
				"GITHUB_API_BASE_URL":        fakeGitHub.URL,
			},
			expErr: `invalid configuration: GITHUB_APP_ID is empty`,
		},
		{
			name: "invalid_private_key_pem",
			env: map[string]string{
				"GITHUB_APP_ID":              "my-app",
				"GITHUB_APP_INSTALLATION_ID": "123",
				"GITHUB_APP_PRIVATE_KEY_PEM": "invalid_pem",
				"GITHUB_PLUGIN_DISPLAY_NAME": testGitHubPluginDisplayName,
				"GITHUB_PLUGIN_HINT":         testGitHubPluginHint,
				"GITHUB_API_BASE_URL":        fakeGitHub.URL,
			},
			expErr: `failed to parse private key`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, done := context.WithCancel(ctx)
			defer done()

			var cmd ServerCommand
			cmd.SetLookupEnv(cli.MultiLookuper(
				cli.MapLookuper(tc.env),
			))

			_, _, _ = cmd.Pipe()

			_, err := cmd.RunUnstarted(ctx, tc.args)
			if diff := testutil.DiffErrString(err, tc.expErr); diff != "" {
				t.Fatal(diff)
			}
			if err != nil {
				return
			}
		})
	}
}
