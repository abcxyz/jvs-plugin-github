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
	"testing"

	jvspb "github.com/abcxyz/jvs/apis/v0"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	testGitHubIssueURL = "https://github.com/test-owner/test-repo/issues/1"
)

type testIssueMatcher struct {
	rPluginGitHubIssue *pluginGitHubIssue
	rErr               error
}

func (t *testIssueMatcher) MatchIssue(ctx context.Context, issueURL string) (*pluginGitHubIssue, error) {
	return t.rPluginGitHubIssue, t.rErr
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		validator *testIssueMatcher
		req       *jvspb.ValidateJustificationRequest
		wantResq  *jvspb.ValidateJustificationResponse
		wantErr   string
	}{
		{
			name: "success",
			validator: &testIssueMatcher{
				rPluginGitHubIssue: &pluginGitHubIssue{
					Owner:       "test-owner",
					RepoName:    "test-repo-name",
					IssueNumber: 1,
				},
				rErr: nil,
			},
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: githubCategory,
					Value:    testGitHubIssueURL,
				},
			},
			wantResq: &jvspb.ValidateJustificationResponse{
				Valid: true,
				Annotation: map[string]string{
					respAnnotationKeyIssueURL:    testGitHubIssueURL,
					respAnnotationKeyIssueOwner:  "test-owner",
					respAnnotationKeyIssueRepo:   "test-repo-name",
					respAnnotationKeyIssueNumber: "1",
				},
			},
		},
		{
			name: "internal_error",
			validator: &testIssueMatcher{
				rPluginGitHubIssue: nil,
				rErr:               fmt.Errorf("injected error"),
			},
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: githubCategory,
					Value:    testGitHubIssueURL,
				},
			},
			wantErr: "injected error",
		},
		{
			name: "wrong_category",
			validator: &testIssueMatcher{
				rPluginGitHubIssue: nil,
				rErr:               nil,
			},
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: "test-category",
					Value:    testGitHubIssueURL,
				},
			},
			wantResq: &jvspb.ValidateJustificationResponse{
				Valid: false,
				Error: []string{`failed to perform validation, expected category "test-category" to be "github"`},
			},
		},
		{
			name: "issue_not_found",
			validator: &testIssueMatcher{
				rPluginGitHubIssue: nil,
				rErr:               errors.Join(errInvalidJustification, fmt.Errorf("issue not found")),
			},
			req: &jvspb.ValidateJustificationRequest{
				Justification: &jvspb.Justification{
					Category: githubCategory,
					Value:    testGitHubIssueURL,
				},
			},
			wantResq: &jvspb.ValidateJustificationResponse{
				Valid: false,
				Error: []string{"invalid justification\nissue not found"},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &GitHubPlugin{
				validator: tc.validator,
			}
			gotResq, gotErr := p.Validate(ctx, tc.req)
			if diff := testutil.DiffErrString(gotErr, tc.wantErr); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tc.wantResq, gotResq, cmpopts.IgnoreUnexported(jvspb.ValidateJustificationResponse{})); diff != "" {
				t.Errorf("Failed validation (-want,+got):\n%s", diff)
			}
		})
	}
}
