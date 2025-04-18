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

package keyutil

import (
	"crypto/rsa"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/abcxyz/pkg/testutil"
)

func TestReadRSAPrivateKey(t *testing.T) {
	t.Parallel()

	testPrivateRSAKeyString, testPrivateRSAKey := TestGenerateRSAPrivateKey(t)

	cases := []struct {
		name        string
		pkPEMString string
		wantPK      *rsa.PrivateKey
		wantErr     string
	}{
		{
			name:        "success",
			pkPEMString: testPrivateRSAKeyString,
			wantPK:      testPrivateRSAKey,
		},
		{
			name:        "invalid_pem",
			pkPEMString: "invalid_format",
			wantErr:     "failed to decode PEM formated key",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotPK, err := ReadRSAPrivateKey(tc.pkPEMString)

			if diff := testutil.DiffErrString(err, tc.wantErr); diff != "" {
				t.Error(diff)
			}
			// if gotPK and wantPK are both nil, no check is needed.
			if !(gotPK == nil && tc.wantPK == nil) {
				// if they are both not nil, compare them.
				if gotPK != nil && tc.wantPK != nil {
					if diff := cmp.Diff(gotPK, tc.wantPK); diff != "" {
						t.Errorf("rsa private key got unexpected diff (-want,+got):\n%s", diff)
					}
				} else {
					t.Errorf("rsa private key got unexpected diff: got %v, want %v", gotPK, tc.wantPK)
				}
			}
		})
	}
}
