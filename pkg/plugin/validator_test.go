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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abcxyz/pkg/githubapp"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-github/v55/github"
)

func testGenerateIssueInfo(tb testing.TB, state string) *github.Issue {
	tb.Helper()
	helperStateString := state
	return &github.Issue{
		State: &helperStateString,
	}
}

func TestCreateValidator(t *testing.T) {
	t.Parallel()

	testPrivateKey := testGeneratePrivateKeyString(t)

	cases := []struct {
		name          string
		cfg           *PluginConfig
		wantErrSubstr string
	}{
		{
			name: "success",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  testPrivateKey,
			},
		},
		{
			name: "invalid_pem",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  "abcde",
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

	testPrivateKey := testGeneratePrivateKeyString(t)

	cases := []struct {
		name                    string
		cfg                     *PluginConfig
		issueURL                string
		issueState              string
		fakeTokenServerResqCode int
		wantErrSubstr           string
	}{
		{
			name:     "success",
			issueURL: "https://github.com/test-owner/test-repo/issues/1",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  testPrivateKey,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			issueState:              "open",
		},
		{
			name:     "invalid_issue_url",
			issueURL: "https://github.com/test-owner/test-repo",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  testPrivateKey,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			wantErrSubstr:           "invalid issue url",
		},
		{
			name:     "issue_not_int",
			issueURL: "https://github.com/test-owner/test-repo/issues/abc",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  testPrivateKey,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			wantErrSubstr:           "failed to convert issueNumber",
		},
		{
			name:     "unauthorized",
			issueURL: "https://github.com/test-owner/test-repo/issues/1",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  testPrivateKey,
			},
			fakeTokenServerResqCode: http.StatusUnauthorized,
			wantErrSubstr:           "failed to get access token",
		},
		{
			name:     "issue_not_open",
			issueURL: "https://github.com/test-owner/test-repo/issues/1",
			cfg: &PluginConfig{
				GithubAppID:             "test-github-id",
				GithubAppInstallationID: "test-install-id",
				GithubAppPrivateKeyPEM:  testPrivateKey,
			},
			fakeTokenServerResqCode: http.StatusCreated,
			issueState:              "closed",
			wantErrSubstr:           "issue is in state: closed",
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

			opts := []githubapp.ConfigOption{githubapp.WithAccessTokenURLPattern(fakeTokenServer.URL + "/%s/access_tokens")}
			validator, err := NewValidator(tc.cfg, WithFakeGithubIssue(testGenerateIssueInfo(t, tc.issueState)))
			if err != nil {
				t.Fatalf("failed to create validator: %v", err)
			}
			gotErr := validator.MatchIssue(ctx, tc.issueURL, opts...)
			if diff := testutil.DiffErrString(gotErr, tc.wantErrSubstr); diff != "" {
				t.Errorf("Process(%+v) got unexpected error substring: %v", tc.name, diff)
			}
		})
	}
}

func testGeneratePrivateKeyString(tb testing.TB) string {
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
	return buf.String()
}
