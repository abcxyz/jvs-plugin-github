// Copyright 2023 Google LLC
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

package plugin

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abcxyz/pkg/githubapp"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-github/v55/github"
)

func TestCreateValidator(t *testing.T) {
	t.Parallel()

	testPrivateKeyString, _ := testGeneratePrivateKey(t)

	cases := []struct {
		name          string
		cfg           *PluginConfig
		wantErrSubstr string
	}{
		{
			name: "success",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
		},
		{
			name: "invalid_pem",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  "abcde",
			},
			wantErrSubstr: "failed to decode PEM formated key",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, gotErr := NewValidator(tc.cfg)
			if diff := testutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected validator creation error substring: %v", tc.name, diff)
			}
		})
	}
}

func TestMatchIssue(t *testing.T) {
	t.Parallel()

	testPrivateKeyString, testPrivateKey := testGeneratePrivateKey(t)

	cases := []struct {
		name                    string
		cfg                     *PluginConfig
		issueURL                string
		issueBytes              []byte
		fakeTokenServerResqCode int
		wantErrSubstr           string
	}{
		{
			name:     "success",
			issueURL: "https://github.com/test-owner/test-repo/issues/1",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			issueBytes:              []byte(`{"state": "open"}`),
		},
		{
			name:     "invalid_issue_url",
			issueURL: "https://github.com/test-owner/test-repo",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			wantErrSubstr:           "invalid issue url",
			issueBytes:              []byte(`{"state": "open"}`),
		},
		{
			name:     "issue_not_int",
			issueURL: "https://github.com/test-owner/test-repo/issues/abc",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			wantErrSubstr:           "invalid issue url, issueURL doesn't match pattern",
			issueBytes:              []byte(`{"state": "open"}`),
		},
		{
			name:     "unauthorized",
			issueURL: "https://github.com/test-owner/test-repo/issues/1",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode: http.StatusUnauthorized,
			wantErrSubstr:           "failed to get access token",
			issueBytes:              []byte(`{"state": "open"}`),
		},
		{
			name:     "issue_not_open",
			issueURL: "https://github.com/test-owner/test-repo/issues/1",
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			wantErrSubstr:           "issue is in state: closed",
			issueBytes:              []byte(`{"state": "closed"}`),
		},
	}

	for _, tc := range cases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Accept") != "application/vnd.github+json" {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing accept header")
					return
				}
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
					w.WriteHeader(500)
					fmt.Fprintf(w, "missing or malformed authorization header")
					return
				}
				w.WriteHeader(tc.fakeTokenServerResqCode)
				fmt.Fprintf(w, `{"token":"this-is-the-token-from-github"}`)
			}))

			hc, done := newTestServer(testHandleIssueReturn(t, tc.issueBytes))
			defer done()
			ghAppOpts := []githubapp.ConfigOption{
				githubapp.WithJWTTokenCaching(1 * time.Minute),
				githubapp.WithAccessTokenURLPattern(fakeTokenServer.URL + "/%s/access_tokens"),
			}
			testGHAppCfg := githubapp.NewConfig(tc.cfg.GitHubAppID, tc.cfg.GitHubAppInstallationID, testPrivateKey, ghAppOpts...)
			testGituhbApp := githubapp.New(testGHAppCfg)

			validator, err := NewValidator(tc.cfg, WithGitHubClient(github.NewClient(hc)), WithGithubApp(testGituhbApp))
			if err != nil {
				t.Fatalf("failed to create validator: %v", err)
			}
			gotErr := validator.MatchIssue(ctx, tc.issueURL)
			if diff := testutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}
		})
	}
}

func testGeneratePrivateKey(tb testing.TB) (string, *rsa.PrivateKey) {
	tb.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("Error generating RSA private key: %v", err)
	}

	// Encode the private key to the PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	buf := new(bytes.Buffer)
	err = pem.Encode(buf, privateKeyPEM)
	if err != nil {
		tb.Fatalf("Error encoding privateKeyPEM: %v", err)
	}
	return buf.String(), privateKey
}

// newTestServer creates a fake http client.
func newTestServer(handler func(w http.ResponseWriter, r *http.Request)) (*http.Client, func()) {
	ts := httptest.NewTLSServer(http.HandlerFunc(handler))
	// Need insecure TLS option for testing.
	// #nosec G402
	tlsConf := &tls.Config{InsecureSkipVerify: true}
	tr := &http.Transport{
		TLSClientConfig: tlsConf,
		DialTLS: func(netw, addr string) (net.Conn, error) {
			return tls.Dial("tcp", ts.Listener.Addr().String(), tlsConf)
		},
	}
	return &http.Client{Transport: tr}, func() {
		tr.CloseIdleConnections()
		ts.Close()
	}
}

// testHandleIssueReturn returns a fake http func that writes the data in http response.
func testHandleIssueReturn(tb testing.TB, data []byte) func(w http.ResponseWriter, r *http.Request) {
	tb.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		switch r.URL.Path {
		case "/repos/test-owner/test-repo/issues/1":
			_, err := w.Write(data)
			if err != nil {
				tb.Fatalf("failed to write response for object info: %v", err)
			}
		default:
			http.Error(w, "injected error", http.StatusNotFound)
		}
	}
}
