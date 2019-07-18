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
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ParseEndpoint is a helper function that reads an endpoint and returns the looked-up IP address
func ParseEndpoint(displayName string, endpoint string) (net.IP, error) {
	// Remove any leading http/https if exists
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	// Remove any trailing / if exists
	endpoint = strings.TrimSuffix(endpoint, "/")

	ip, err := net.LookupIP(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"display_name": displayName,
			"endpoint":     endpoint,
		}).Error("Could not parse endpoint into IP")
		return nil, err
	}

	// Return IPv4
	return ip[0], nil
}
