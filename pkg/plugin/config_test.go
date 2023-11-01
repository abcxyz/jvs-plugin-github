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

// Package plugin provides the implementation of the JVS plugin interface.
package plugin

import (
	"testing"

	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

const (
	testGitHubAppID             = "123456"
	testGitHubAppInstallationID = "12345678"
)

func TestPluginConfig_ToFlags(t *testing.T) {
	t.Parallel()

	testPrivateKeyString, _ := testGeneratePrivateKey(t)

	cases := []struct {
		name       string
		args       []string
		envs       map[string]string
		wantConfig *PluginConfig
	}{
		{
			name: "all_envs_specified",
			envs: map[string]string{
				"GITHUB_APP_ID":              testGitHubAppID,
				"GITHUB_APP_INSTALLATION_ID": testGitHubAppInstallationID,
				"GITHUB_APP_PRIVATE_KEY_PEM": testPrivateKeyString,
			},
			wantConfig: &PluginConfig{
				GitHubAppID:             testGitHubAppID,
				GitHubAppInstallationID: testGitHubAppInstallationID,
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
		},
		{
			name: "all_flags_specified",
			args: []string{
				"-github-app-id", testGitHubAppID,
				"-github-app-installation-id", testGitHubAppInstallationID,
				"-github-app-private-key-pem", testPrivateKeyString,
			},
			wantConfig: &PluginConfig{
				GitHubAppID:             testGitHubAppID,
				GitHubAppInstallationID: testGitHubAppInstallationID,
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotConfig := &PluginConfig{}
			set := cli.NewFlagSet(cli.WithLookupEnv(cli.MapLookuper(tc.envs)))
			set = gotConfig.ToFlags(set)
			if err := set.Parse(tc.args); err != nil {
				t.Errorf("unexpected flag set parse error: %v", err)
			}
			if diff := cmp.Diff(tc.wantConfig, gotConfig); diff != "" {
				t.Errorf("Config unexpected diff (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestPluginConfig_Validate(t *testing.T) {
	t.Parallel()

	testPrivateKeyString, _ := testGeneratePrivateKey(t)

	cases := []struct {
		name    string
		cfg     *PluginConfig
		wantErr string
	}{
		{
			name: "success",
			cfg: &PluginConfig{
				GitHubAppID:             testGitHubAppID,
				GitHubAppInstallationID: testGitHubAppInstallationID,
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
		},
		{
			name: "empty_github_app_id",
			cfg: &PluginConfig{
				GitHubAppID:             "",
				GitHubAppInstallationID: testGitHubAppInstallationID,
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			wantErr: "GITHUB_APP_ID is empty",
		},
		{
			name: "empty_github_app_installation_id",
			cfg: &PluginConfig{
				GitHubAppID:             testGitHubAppID,
				GitHubAppInstallationID: "",
				GitHubAppPrivateKeyPEM:  testPrivateKeyString,
			},
			wantErr: "GITHUB_APP_INSTALLATION_ID is empty",
		},
		{
			name: "empty_github_app_private_key_pem",
			cfg: &PluginConfig{
				GitHubAppID:             testGitHubAppID,
				GitHubAppInstallationID: testGitHubAppInstallationID,
				GitHubAppPrivateKeyPEM:  "",
			},
			wantErr: "GITHUB_APP_PRIVATE_KEY_PEM is empty",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Errorf("Unexpected err: %s", diff)
			}
		})
	}
}
