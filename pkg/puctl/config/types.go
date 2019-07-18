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

package config

// Establish some well known config keys and default values
const (
	PodCIDRKey                  = "global.podCIDR"
	ServiceCIDRKey              = "global.serviceCIDR"
	Pure1UnpluggedInstallDirKey = "global.installDir"
	PublicAddressKey            = "global.publicAddress"
	SSLCertFileKey              = "global.sslCertFile"
	SSLKeyFileKey               = "global.sslKeyFile"
	CreateSelfSignedCertsKey    = "global.createSelfSignedCerts"
)

// Defaults is a map of config key to default value
var Defaults = map[string]string{
	PodCIDRKey:                  "192.168.0.0/16",
	ServiceCIDRKey:              "10.96.0.0/12",
	Pure1UnpluggedInstallDirKey: "/opt/pure1-unplugged",
	PublicAddressKey:            "pure1-unplugged.example.com",
	SSLCertFileKey:              "/etc/pure1-unplugged/ssl/pure1-unplugged.crt",
	SSLKeyFileKey:               "/etc/pure1-unplugged/ssl/pure1-unplugged.key",
	CreateSelfSignedCertsKey:    "true",
}
