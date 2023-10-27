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

// PluginConfig defines the set over environment variables required
// for running the plugin.
type PluginConfig struct {
	// ID of the GitHub APP we use to authenticate.
	GitHubAppID string `env:"GITHUB_APP_ID,required"`
	// Installation ID of the github app.
	GitHubAppInstallationID string `env:"GITHUB_APP_INSTALLATION_ID,required"`
	// The private Key PEM obtained for github app.
	GitHubAppPrivateKeyPEM string `env:"GITHUB_APP_PRIVATE_KEY_PEM,required"`
}
