// Copyright 2019, Pure Storage Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oidc

import (
	"context"

	oidc "github.com/coreos/go-oidc"
)

// NewOIDCVerifier constructs a new oidc.Verifier with the given IDTokenVerifier
func NewOIDCVerifier(v *oidc.IDTokenVerifier) Verifier {
	return Verifier{verifier: v}
}

// Verify checks that a given ID token is valid
func (o *Verifier) Verify(ctx context.Context, rawIDToken string) error {
	_, err := o.verifier.Verify(ctx, rawIDToken)
	return err
}
