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

// Package keyutil provides commonly used test functions to generate RSA keys.
package keyutil

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// ReadRSAPrivateKey encrypts a PEM encoding RSA key string and returns the decoded RSA key.
func ReadRSAPrivateKey(rsaPrivateKeyPEM string) (*rsa.PrivateKey, error) {
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

// TestGenerateRsaPrivateKey generates a rsa Key for testing use.
// It returns the PEM decoded private key string and the rsa.PrivateKey it itself.
func TestGenerateRsaPrivateKey(tb testing.TB) (string, *rsa.PrivateKey) {
	tb.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("failed to generate rsa private key: %v", err)
	}

	// Encode the private key to the PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	var buf bytes.Buffer
	if err = pem.Encode(&buf, privateKeyPEM); err != nil {
		tb.Fatalf("failed to encode to pem: %v", err)
	}
	return buf.String(), privateKey
}
