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
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abcxyz/pkg/githubapp"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v55/github"
)

const (
	issueURLHost            = "https://github.com"
	testIssueOwner          = "test-owner"
	testIssueRepoName       = "test-repo"
	testExistIssueNumber    = 1
	testNonExistIssueNumber = 2
	issueRESTAPIPathPrefix  = "/repos"
)

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
		wantPluginGitHubIssue   *pluginGitHubIssue
		// check is returned error is the correct type
		isInvalidJustificationErr bool
	}{
		{
			name:     "success",
			issueURL: fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testExistIssueNumber),
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			issueBytes:              []byte(`{"state": "open"}`),
			wantPluginGitHubIssue: &pluginGitHubIssue{
				Owner:       testIssueOwner,
				RepoName:    testIssueRepoName,
				IssueNumber: testExistIssueNumber,
			},
		},
		{
			name:     "invalid_issue_url",
			issueURL: fmt.Sprintf("%s/%s/%s", issueURLHost, testIssueOwner, testIssueRepoName),
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode:   http.StatusCreated,
			wantErrSubstr:             "invalid issue url",
			isInvalidJustificationErr: true,
			issueBytes:                []byte(`{"state": "open"}`),
		},
		{
			name:     "issue_not_int",
			issueURL: fmt.Sprintf("%s/%s/%s/issues/%s", issueURLHost, testIssueOwner, testIssueRepoName, "abc"),
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode:   http.StatusCreated,
			wantErrSubstr:             "invalid issue url, issueURL doesn't match pattern",
			isInvalidJustificationErr: true,
			issueBytes:                []byte(`{"state": "open"}`),
		},
		{
			name:     "unauthorized",
			issueURL: fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testExistIssueNumber),
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode:   http.StatusUnauthorized,
			wantErrSubstr:             "failed to get access token",
			isInvalidJustificationErr: false,
			issueBytes:                []byte(`{"state": "open"}`),
		},
		{
			name:     "issue_not_open",
			issueURL: fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testExistIssueNumber),
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode:   http.StatusCreated,
			wantErrSubstr:             "issue is in state: closed",
			isInvalidJustificationErr: true,
			issueBytes:                []byte(`{"state": "closed"}`),
			wantPluginGitHubIssue: &pluginGitHubIssue{
				Owner:       testIssueOwner,
				RepoName:    testIssueRepoName,
				IssueNumber: testExistIssueNumber,
			},
		},
		{
			name:     "issue_not_exist",
			issueURL: fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testNonExistIssueNumber),
			cfg: &PluginConfig{
				GitHubAppID:             "test-github-id",
				GitHubAppInstallationID: "test-install-id",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			fakeTokenServerResqCode:   http.StatusCreated,
			wantErrSubstr:             "issue not found",
			isInvalidJustificationErr: true,
			issueBytes:                []byte(`{"state": "closed"}`),
			wantPluginGitHubIssue: &pluginGitHubIssue{
				Owner:       testIssueOwner,
				RepoName:    testIssueRepoName,
				IssueNumber: testNonExistIssueNumber,
			},
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
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "missing or malformed authorization header")
					return
				}
				w.WriteHeader(tc.fakeTokenServerResqCode)
				fmt.Fprintf(w, `{"token":"this-is-the-token-from-github"}`)
			}))
			t.Cleanup(func() {
				fakeTokenServer.Close()
			})

			hc := newTestServer(t, testHandleIssueReturn(t, tc.issueBytes))
			testGitHubClient := github.NewClient(hc)

			ghAppOpts := []githubapp.ConfigOption{
				githubapp.WithJWTTokenCaching(1 * time.Minute),
				githubapp.WithAccessTokenURLPattern(fakeTokenServer.URL + "/%s/access_tokens"),
			}
			testGHAppCfg := githubapp.NewConfig(tc.cfg.GitHubAppID, tc.cfg.GitHubAppInstallationID, testPrivateKey, ghAppOpts...)
			testGithubApp := githubapp.New(testGHAppCfg)

			validator := NewValidator(testGitHubClient, testGithubApp)
			gotPluginGitHubIssue, gotErr := validator.MatchIssue(ctx, tc.issueURL)
			if diff := testutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}
			if diff := cmp.Diff(gotPluginGitHubIssue, tc.wantPluginGitHubIssue); diff != "" {
				t.Errorf("Process(%+v) got unexpected pluginGitHubIssue diff (-want, +got):\n%s", tc.name, diff)
			}
			if tc.wantErrSubstr != "" {
				if tc.isInvalidJustificationErr {
					if !errors.Is(gotErr, errInvalidJustification) {
						t.Errorf("Process(%+v) got unexpected error type, expect error to be of type: %v", tc.name, errInvalidJustification)
					}
				} else {
					if errors.Is(gotErr, errInvalidJustification) {
						t.Errorf("Process(%+v) got unexpected error type, expect error NOT to be of type: %v", tc.name, errInvalidJustification)
					}
				}
			}
		})
	}
}

// testGeneratePrivateKey generates a rsa Key for testing use.
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
func newTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *http.Client {
	t.Helper()

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

	t.Cleanup(func() {
		tr.CloseIdleConnections()
		ts.Close()
	})
	return &http.Client{Transport: tr}
}

// testHandleIssueReturn returns a fake http func that writes the data in http response.
func testHandleIssueReturn(tb testing.TB, data []byte) func(w http.ResponseWriter, r *http.Request) {
	tb.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case fmt.Sprintf("%s/%s/%s/issues/%v", issueRESTAPIPathPrefix, testIssueOwner, testIssueRepoName, testExistIssueNumber):
			if _, err := w.Write(data); err != nil {
				tb.Fatalf("failed to write response for object info: %v", err)
			}
		case fmt.Sprintf("%s/%s/%s/issues/%v", issueRESTAPIPathPrefix, testIssueOwner, testIssueRepoName, testNonExistIssueNumber):
			http.Error(w, "issue not found", http.StatusNotFound)
		default:
			http.Error(w, "injected server error", http.StatusInternalServerError)
		}
	}
}
