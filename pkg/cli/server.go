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

// Package cli implements the commands for the github plugin CLI.
package cli

import (
	"context"
	"fmt"

	"github.com/google/go-github/v55/github"
	goplugin "github.com/hashicorp/go-plugin"

	"github.com/abcxyz/jvs-plugin-github/pkg/plugin"
	jvspb "github.com/abcxyz/jvs/apis/v0"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/logging"
)

type ServerCommand struct {
	cli.BaseCommand

	cfg *plugin.PluginConfig
}

func (c *ServerCommand) Desc() string {
	return `Start GitHub Plugin`
}

func (c *ServerCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

  Start a GitHub Plugin.
`
}

func (c *ServerCommand) Flags() *cli.FlagSet {
	c.cfg = &plugin.PluginConfig{}
	set := c.NewFlagSet()
	return c.cfg.ToFlags(set)
}

func (c *ServerCommand) Run(ctx context.Context, args []string) error {
	p, err := c.RunUnstarted(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to instantiate github plugin: %w", err)
	}

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: jvspb.Handshake,
		Plugins: map[string]goplugin.Plugin{
			"jvs-plugin-github": &jvspb.ValidatorPlugin{Impl: p},
		},

		// A non-nil value here enables gRPC serving for this plugin.
		GRPCServer: goplugin.DefaultGRPCServer,
	})

	return nil
}

func (c *ServerCommand) RunUnstarted(ctx context.Context, args []string) (*plugin.GitHubPlugin, error) {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}
	args = f.Args()
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %q", args)
	}

	logger := logging.FromContext(ctx)

	if err := c.cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	logger.DebugContext(ctx, "loaded configuration",
		"github_app_id", c.cfg.GitHubAppID,
		"github_app_installation_id", c.cfg.GitHubAppInstallationID)

	//  If a nil httpClient is provided, a new http.Client will be used.
	ghClient := github.NewClient(nil)

	signer, err := githubauth.NewPrivateKeySigner(c.cfg.GitHubAppPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	ghApp, err := githubauth.NewApp(c.cfg.GitHubAppID, signer,
		githubauth.WithBaseURL(c.cfg.GitHubAPIBaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}

	ghInstall, err := ghApp.InstallationForID(ctx, c.cfg.GitHubAppInstallationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get github installation: %w", err)
	}

	p := plugin.NewGitHubPlugin(ctx, ghClient, ghInstall, c.cfg)
	return p, nil
}
