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
	"crypto/rsa"
	"fmt"

	"github.com/abcxyz/jvs-plugin-github/pkg/plugin"
	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/githubapp"
	"github.com/abcxyz/pkg/logging"
	"github.com/google/go-github/v55/github"
	"github.com/lestrrat-go/jwx/v2/jwk"

	jvspb "github.com/abcxyz/jvs/apis/v0"
	goplugin "github.com/hashicorp/go-plugin"
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
	logger.DebugContext(ctx, "loaded configuration", "config", c.cfg)

	//  If a nil httpClient is provided, a new http.Client will be used.
	ghClient := github.NewClient(nil)
	pk, err := readPrivateKey(c.cfg.GitHubAppPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("invalid private key pem: %w", err)
	}
	ghAppCfg := githubapp.NewConfig(c.cfg.GitHubAppID, c.cfg.GitHubAppInstallationID, pk)
	ghApp := githubapp.New(ghAppCfg)

	p := plugin.NewGitHubPlugin(ctx, ghClient, ghApp)
	return p, nil
}

// readPrivateKey encrypts a PEM encoding RSA key string and returns the decoded RSA key.
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
