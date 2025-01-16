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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v55/github"

	"github.com/abcxyz/jvs-plugin-github/pkg/plugin/keyutil"
	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/testutil"
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

	cases := []struct {
		name                    string
		issueURL                string
		issueBytes              []byte
		fakeTokenServerResqCode int
		wantErrSubstr           string
		wantPluginGitHubIssue   *pluginGitHubIssue
		// check is returned error is the correct type
		isInvalidJustificationErr bool
	}{
		{
			name:                    "success",
			issueURL:                fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testExistIssueNumber),
			fakeTokenServerResqCode: http.StatusCreated,
			issueBytes:              []byte(`{"state": "open"}`),
			wantPluginGitHubIssue: &pluginGitHubIssue{
				Owner:       testIssueOwner,
				RepoName:    testIssueRepoName,
				IssueNumber: testExistIssueNumber,
			},
		},
		{
			name:                      "invalid_issue_url",
			issueURL:                  fmt.Sprintf("%s/%s/%s", issueURLHost, testIssueOwner, testIssueRepoName),
			fakeTokenServerResqCode:   http.StatusCreated,
			wantErrSubstr:             "invalid issue url",
			isInvalidJustificationErr: true,
			issueBytes:                []byte(`{"state": "open"}`),
		},
		{
			name:                      "issue_not_int",
			issueURL:                  fmt.Sprintf("%s/%s/%s/issues/%s", issueURLHost, testIssueOwner, testIssueRepoName, "abc"),
			fakeTokenServerResqCode:   http.StatusCreated,
			wantErrSubstr:             "invalid issue url, issueURL doesn't match pattern",
			isInvalidJustificationErr: true,
			issueBytes:                []byte(`{"state": "open"}`),
		},
		{
			name:                      "unauthorized",
			issueURL:                  fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testExistIssueNumber),
			fakeTokenServerResqCode:   http.StatusUnauthorized,
			wantErrSubstr:             "failed to get access token",
			isInvalidJustificationErr: false,
			issueBytes:                []byte(`{"state": "open"}`),
		},
		{
			name:                      "issue_not_open",
			issueURL:                  fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testExistIssueNumber),
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
			name:                      "issue_not_exist",
			issueURL:                  fmt.Sprintf("%s/%s/%s/issues/%v", issueURLHost, testIssueOwner, testIssueRepoName, testNonExistIssueNumber),
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			fakeGitHub := func() *httptest.Server {
				mux := http.NewServeMux()
				mux.Handle("GET /app/installations/123", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `{"access_tokens_url": "http://%s/app/installations/123/access_tokens"}`, r.Host)
				}))
				mux.Handle("POST /app/installations/123/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.fakeTokenServerResqCode)
					fmt.Fprintf(w, `{"token": "this-is-the-token-from-github"}`)
				}))
				return httptest.NewServer(mux)
			}()
			t.Cleanup(func() {
				fakeGitHub.Close()
			})

			_, testPrivateKey := keyutil.TestGenerateRSAPrivateKey(t)

			hc := newTestServer(t, testHandleIssueReturn(t, tc.issueBytes))
			testGitHubClient := github.NewClient(hc)

			testGitHubApp, err := githubauth.NewApp("my-app", testPrivateKey,
				githubauth.WithBaseURL(fakeGitHub.URL))
			if err != nil {
				t.Fatal(err)
			}

			installation, err := testGitHubApp.InstallationForID(ctx, "123")
			if err != nil {
				t.Fatal(err)
			}

			validator := NewValidator(testGitHubClient, installation)
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
