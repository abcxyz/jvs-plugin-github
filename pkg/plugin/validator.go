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
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/abcxyz/pkg/githubapp"
	"github.com/google/go-github/v55/github"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// Validator validates github issue against validation criteria.
type Validator struct {
	cfg        *PluginConfig
	decodedPem *rsa.PrivateKey
	client     *github.Client

	// for testing only
	fakeGithubIssue *github.Issue
}

// ExchangeResponse is the GitHub API response of requesting an access token
// for the GitHub App installation with requested repositories and permissions.
type ExchangeResponse struct {
	AccessToken string `json:"token"`
}

// pluginGithubIssue contains the required attribute parsed from
// the issue URL.
type pluginGithubIssue struct {
	Owner       string
	RepoName    string
	IssueNumber int
}

// NewValidator creates a validator.
func NewValidator(cfg *PluginConfig) (*Validator, error) {
	pk, err := readPrivateKey(cfg.GitHubAppPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	v := Validator{
		cfg:        cfg,
		decodedPem: pk,
		client:     github.NewClient(nil),
	}
	return &v, nil
}

// MatchIssue parses issue info from provided issueURL and validate if the issue is valid.
func (v *Validator) MatchIssue(ctx context.Context, issueURL string, opts ...githubapp.ConfigOption) error {
	info, err := parseIssueInfoFromURL(issueURL)
	if err != nil {
		return fmt.Errorf("failed to parse issueURL: %w", err)
	}

	t, err := v.getAccessToken(ctx, v.cfg, v.decodedPem, info.RepoName, opts)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	v.client = v.client.WithAuthToken(t)

	return v.validateIssue(ctx, info)
}

// validateIssue verifies if the issue exists and the issue is open.
func (v *Validator) validateIssue(ctx context.Context, pi *pluginGithubIssue) error {
	issue, err := v.getGithubIssue(ctx, pi)
	if err != nil {
		return fmt.Errorf("failed to get issue info: %w", err)
	}
	if s := issue.GetState(); s != "open" {
		return fmt.Errorf("issue is in state: %s, please make sure to use an open issue", s)
	}
	return nil
}

// getGithubIssue gets the provided issue's info from github api.
func (v *Validator) getGithubIssue(ctx context.Context, pi *pluginGithubIssue) (*github.Issue, error) {
	if v.fakeGithubIssue != nil {
		return v.fakeGithubIssue, nil
	}
	issue, _, err := v.client.Issues.Get(ctx, pi.Owner, pi.RepoName, pi.IssueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// getAccessToken gets an access token with issue read permission to the repo
// which contains the issue.
func (v *Validator) getAccessToken(ctx context.Context, cfg *PluginConfig, pk *rsa.PrivateKey, repoName string, opts []githubapp.ConfigOption) (string, error) {
	opts = append(opts, githubapp.WithJWTTokenCaching(1*time.Minute))

	ghCfg := githubapp.NewConfig(cfg.GitHubAppID, cfg.GitHubAppInstallationID, pk, opts...)
	githubApp := githubapp.New(ghCfg)

	tr := githubapp.TokenRequest{
		Repositories: []string{repoName},
		Permissions: map[string]string{
			"issues": "read",
		},
	}
	var tokenResp ExchangeResponse
	resp, err := githubApp.AccessToken(ctx, &tr)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	if err := json.Unmarshal([]byte(resp), &tokenResp); err != nil {
		return "", fmt.Errorf("error unmarshal resp: %w", err)
	}
	return tokenResp.AccessToken, nil
}

// parseGithubIssue parses issue info from Issue URL.
func parseIssueInfoFromURL(issueURL string) (*pluginGithubIssue, error) {
	u, err := url.Parse(issueURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provided issue url: %w", err)
	}
	arr := strings.Split(u.Path, "/")
	if len(arr) != 5 || arr[3] != "issues" {
		return nil, fmt.Errorf("invalid issue url, please check your issues url is in format https://github.com/<org_name>/<repo_id>/issues/<issue_number>")
	}

	issueNumber, err := strconv.Atoi(arr[4])
	if err != nil {
		return nil, fmt.Errorf("failed to convert issueNumber %s to int: %w", arr[4], err)
	}

	return &pluginGithubIssue{
		Owner:       arr[1],
		RepoName:    arr[2],
		IssueNumber: issueNumber,
	}, nil
}

// readPrivateKey reads a RSA encrypted private key using PEM encoding as a string
// and returns an RSA key.
func readPrivateKey(rsaPrivateKeyPEM string) (*rsa.PrivateKey, error) {
	parsedKey, _, err := jwk.DecodePEM([]byte(rsaPrivateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PEM formated key:  %w", err)
	}
	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to convert to *rsa.PrivateKey (got %T)", parsedKey)
	}
	return privateKey, nil
}
