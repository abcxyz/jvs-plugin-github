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

// Package testhelper provides commonly used test functions in different packages.
package testhelper

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// TestGeneratePrivateKey generates a rsa Key for testing use.
// It returns the PEM decoded private key string and the rsa.PrivateKey it itself.
func TestGeneratePrivateKey(tb testing.TB) (string, *rsa.PrivateKey) {
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
