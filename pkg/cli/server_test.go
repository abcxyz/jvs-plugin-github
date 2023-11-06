// Copyright 2023 The Authors (see AUTHORS file)
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

package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/abcxyz/pkg/cli"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/testutil"
)

const (
	testGitHubPluginDisplayName = "test DisplayName"
	testGitHubPluginHint        = "test Hint"
)

func TestServerCommand(t *testing.T) {
	t.Parallel()

	testPrivateKeyString, _ := testGeneratePrivateKey(t)

	ctx := logging.WithLogger(context.Background(), logging.TestLogger(t))

	cases := []struct {
		name   string
		args   []string
		env    map[string]string
		expErr string
	}{
		{
			name: "success",
			env: map[string]string{
				"GITHUB_APP_ID":              "123456",
				"GITHUB_APP_INSTALLATION_ID": "123456",
				"GITHUB_APP_PRIVATE_KEY_PEM": testPrivateKeyString,
				"GITHUB_PLUGIN_DISPLAY_NAME": testGitHubPluginDisplayName,
				"GITHUB_PLUGIN_HINT":         testGitHubPluginHint,
			},
		},
		{
			name:   "unexpected args",
			args:   []string{"foo"},
			expErr: `unexpected arguments: ["foo"]`,
		},
		{
			name: "invalid_config_missing_github_app_id",
			env: map[string]string{
				"GITHUB_APP_INSTALLATION_ID": "123456",
				"GITHUB_APP_PRIVATE_KEY_PEM": testPrivateKeyString,
				"GITHUB_PLUGIN_DISPLAY_NAME": testGitHubPluginDisplayName,
				"GITHUB_PLUGIN_HINT":         testGitHubPluginHint,
			},
			expErr: `invalid configuration: GITHUB_APP_ID is empty`,
		},
		{
			name: "invalid_private_key_pem",
			env: map[string]string{
				"GITHUB_APP_ID":              "123456",
				"GITHUB_APP_INSTALLATION_ID": "123456",
				"GITHUB_APP_PRIVATE_KEY_PEM": "invalid_pem",
				"GITHUB_PLUGIN_DISPLAY_NAME": testGitHubPluginDisplayName,
				"GITHUB_PLUGIN_HINT":         testGitHubPluginHint,
			},
			expErr: `failed to decode PEM formated key`,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, done := context.WithCancel(ctx)
			defer done()

			var cmd ServerCommand
			cmd.SetLookupEnv(cli.MultiLookuper(
				cli.MapLookuper(tc.env),
			))

			_, _, _ = cmd.Pipe()

			_, err := cmd.RunUnstarted(ctx, tc.args)
			if diff := testutil.DiffErrString(err, tc.expErr); diff != "" {
				t.Fatal(diff)
			}
			if err != nil {
				return
			}
		})
	}
}

// testGeneratePrivateKey generates a rsa Key for testing use.
func testGeneratePrivateKey(tb testing.TB) (string, *rsa.PrivateKey) {
	tb.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("Error generating RSA private key: %v", err)
	}

	// Encode the private key to the PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	var buf bytes.Buffer
	if err = pem.Encode(&buf, privateKeyPEM); err != nil {
		tb.Fatalf("Error encoding privateKeyPEM: %v", err)
	}
	return buf.String(), privateKey
}
