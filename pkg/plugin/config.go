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
	"errors"
	"fmt"

	"github.com/abcxyz/pkg/cli"
)

// PluginConfig defines the set over environment variables required
// for running the plugin.
type PluginConfig struct {
	// ID of the GitHub APP we use to authenticate.
	GitHubAppID string
	// Installation ID of the github app.
	GitHubAppInstallationID string
	// The private Key PEM obtained for github app.
	GitHubAppPrivateKeyPEM string
}

// Validate validates if the config is valid.
func (cfg *PluginConfig) Validate() error {
	var rErr error
	if cfg.GitHubAppID == "" {
		rErr = errors.Join(rErr, fmt.Errorf("GITHUB_APP_ID is empty"))
	}
	if cfg.GitHubAppInstallationID == "" {
		rErr = errors.Join(rErr, fmt.Errorf("GITHUB_APP_INSTALLATION_ID is empty"))
	}
	if cfg.GitHubAppPrivateKeyPEM == "" {
		rErr = errors.Join(rErr, fmt.Errorf("GITHUB_APP_PRIVATE_KEY_PEM is empty"))
	}

	return rErr
}

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *PluginConfig) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	// Command options
	f := set.NewSection("GITHUB PLUGIN OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "github-app-id",
		Target:  &cfg.GitHubAppID,
		EnvVar:  "GITHUB_APP_ID",
		Example: "111111",
		Usage:   "The ID of the github app.",
	})

	f.StringVar(&cli.StringVar{
		Name:    "github-app-installation-id",
		Target:  &cfg.GitHubAppInstallationID,
		EnvVar:  "GITHUB_APP_INSTALLATION_ID",
		Example: "project = JRA and assignee != jsmith",
		Usage:   "The installation ID of the github app.",
	})

	f.StringVar(&cli.StringVar{
		Name:   "github-app-private-key-pem",
		Target: &cfg.GitHubAppPrivateKeyPEM,
		EnvVar: "GITHUB_APP_PRIVATE_KEY_PEM",
		Usage:  "The private key pem obtained for github app.",
	})
	return set
}
