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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/abcxyz/pkg/githubapp"
	"github.com/google/go-github/v55/github"
)

const (
	issueURLPatternRegExp = `^https:\/\/github.com\/([a-zA-Z0-9-]*)\/[a-zA-Z0-9-]*\/issues\/[0-9]+$`
)

// Validator validates github issue against validation criteria.
type Validator struct {
	client    *github.Client
	githubApp *githubapp.GitHubApp
}

// ExchangeResponse is the GitHub API response of requesting an access token
// for the GitHub App installation with requested repositories and permissions.
type ExchangeResponse struct {
	AccessToken string `json:"token"`
}

// pluginGitHubIssue contains the required attribute parsed from
// the issue URL.
type pluginGitHubIssue struct {
	Owner       string
	RepoName    string
	IssueNumber int
}

// NewValidator creates a validator.
func NewValidator(ghClinet *github.Client, ghApp *githubapp.GitHubApp) *Validator {
	return &Validator{
		client:    ghClinet,
		githubApp: ghApp,
	}
}

// MatchIssue parses issue info from provided issueURL and validate if the issue is valid.
func (v *Validator) MatchIssue(ctx context.Context, issueURL string) (*pluginGitHubIssue, error) {
	info, err := parseIssueInfoFromURL(issueURL)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse issueURL: %w", errInvalidJustification, err)
	}

	t, err := v.getAccessToken(ctx, info.RepoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	v.client = v.client.WithAuthToken(t)

	return info, v.validateIssue(ctx, info)
}

// validateIssue verifies if the issue exists and the issue is open.
func (v *Validator) validateIssue(ctx context.Context, pi *pluginGitHubIssue) error {
	issue, resp, err := v.client.Issues.Get(ctx, pi.Owner, pi.RepoName, pi.IssueNumber)
	if err != nil {
		// When the issue doesn't not exist, github rest api will return a 404
		// all other non-200 status code will be treated as internal error.
		//
		// See: https://docs.github.com/en/rest/issues/issues?apiVersion=2022-11-28#get-an-issue--status-codes.
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("%w: issue not found: %w", errInvalidJustification, err)
		}
		return fmt.Errorf("failed to get issue info: %w", err)
	}
	if s := issue.GetState(); s != "open" {
		return fmt.Errorf("%w: issue is in state: %s, please make sure to use an open issue", errInvalidJustification, s)
	}
	return nil
}

// getAccessToken gets an access token with issue read permission to the repo
// which contains the issue.
func (v *Validator) getAccessToken(ctx context.Context, repoName string) (string, error) {
	tr := &githubapp.TokenRequest{
		Repositories: []string{repoName},
		Permissions: map[string]string{
			"issues": "read",
		},
	}

	resp, err := v.githubApp.AccessToken(ctx, tr)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	var tokenResp ExchangeResponse
	if err := json.Unmarshal([]byte(resp), &tokenResp); err != nil {
		return "", fmt.Errorf("error unmarshal resp: %w", err)
	}
	return tokenResp.AccessToken, nil
}

// parseIssueInfoFromURL parses pluginGitHubIssue from Issue URL.
func parseIssueInfoFromURL(issueURL string) (*pluginGitHubIssue, error) {
	if match, _ := regexp.MatchString(issueURLPatternRegExp, issueURL); !match {
		return nil, fmt.Errorf("invalid issue url, issueURL doesn't match pattern: %s", issueURLPatternRegExp)
	}
	u, err := url.Parse(issueURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provided issue url: %w", err)
	}

	arr := strings.Split(u.Path, "/")
	// len(arr) is not checked here as regexp.MatchString already covers this.
	issueNumber, err := strconv.Atoi(arr[4])
	if err != nil {
		return nil, fmt.Errorf("failed to convert issueNumber %s to int: %w", arr[4], err)
	}

	return &pluginGitHubIssue{
		Owner:       arr[1],
		RepoName:    arr[2],
		IssueNumber: issueNumber,
	}, nil
}
