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

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DemoIP = "216.58.195.238" // google.com
)

func TestParseEndpointHTTPSWithSlash(t *testing.T) {
	ip, err := ParseEndpoint("test-client", fmt.Sprintf("https://%s/", DemoIP))
	assert.NoError(t, err)
	assert.Equal(t, DemoIP, ip.String())
}

func TestParseEndpointPHTTP(t *testing.T) {
	ip, err := ParseEndpoint("test-client", fmt.Sprintf("http://%s", DemoIP))
	assert.NoError(t, err)
	assert.Equal(t, DemoIP, ip.String())
}
