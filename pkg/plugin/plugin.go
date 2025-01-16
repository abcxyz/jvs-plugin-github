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
	"errors"
	"fmt"
	"strconv"

	"github.com/google/go-github/v55/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jvspb "github.com/abcxyz/jvs/apis/v0"
	"github.com/abcxyz/pkg/githubauth"
)

const (
	// githubCategory is the justification category this plugin will be validating.
	githubCategory               = "github"
	respAnnotationKeyIssueURL    = "github_issue_url"
	respAnnotationKeyIssueOwner  = "github_issue_owner"
	respAnnotationKeyIssueRepo   = "github_issue_repo"
	respAnnotationKeyIssueNumber = "github_issue_number"
)

// issueMatcher is the mockable interface for the convenience of testing.
type issueMatcher interface {
	MatchIssue(ctx context.Context, issueURL string) (*pluginGitHubIssue, error)
}

// GitHubPlugin is the implementation of jvspb.Validator interface.
//
// See: https://pkg.go.dev/github.com/abcxyz/jvs@v0.1.4/apis/v0#Validator
type GitHubPlugin struct {
	// validator implements issueMatcher for validating github issues.
	validator issueMatcher
	// uiData contains the data for ui to display
	uiData *jvspb.UIData
}

// NewGitHubPlugin creates a new GitHubPlugin.
func NewGitHubPlugin(ctx context.Context, ghClient *github.Client, ghInstall *githubauth.AppInstallation, cfg *PluginConfig) *GitHubPlugin {
	return &GitHubPlugin{
		validator: NewValidator(ghClient, ghInstall),
		uiData: &jvspb.UIData{
			DisplayName: cfg.GitHubPluginDisplayName,
			Hint:        cfg.GitHubPluginHint,
		},
	}
}

// Validate returns the validation result.
func (g *GitHubPlugin) Validate(ctx context.Context, req *jvspb.ValidateJustificationRequest) (*jvspb.ValidateJustificationResponse, error) {
	if got, want := req.GetJustification().GetCategory(), githubCategory; got != want {
		return generateInvalidErrResq(fmt.Sprintf("failed to perform validation, expected category %q to be %q", got, want)), nil
	}

	info, err := g.validator.MatchIssue(ctx, req.GetJustification().GetValue())
	if err != nil {
		if errors.Is(err, errInvalidJustification) {
			return generateInvalidErrResq(err.Error()), nil
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &jvspb.ValidateJustificationResponse{
		Valid: true,
		Annotation: map[string]string{
			respAnnotationKeyIssueURL:    req.GetJustification().GetValue(),
			respAnnotationKeyIssueOwner:  info.Owner,
			respAnnotationKeyIssueRepo:   info.RepoName,
			respAnnotationKeyIssueNumber: strconv.Itoa(info.IssueNumber),
		},
	}, nil
}

// GetUIData returns UIData for jvs ui service to use.
func (g *GitHubPlugin) GetUIData(ctx context.Context, req *jvspb.GetUIDataRequest) (*jvspb.UIData, error) {
	return g.uiData, nil
}

// generateInvalidErrResq generates a ValidateJustificationResponse indicating
// the justification is invalid, and use the provided string to set Error field.
func generateInvalidErrResq(s string) *jvspb.ValidateJustificationResponse {
	return &jvspb.ValidateJustificationResponse{
		Valid: false,
		Error: []string{s},
	}
}
