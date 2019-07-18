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

package array

import (
	"fmt"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/array/flasharray"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/array/flashblade"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
)

// Type guard: check that this struct implements the interface
var _ resources.CollectorFactory = (*restFactory)(nil)

// NewRESTFactory produces a RESTFactory, which produces REST client Collectors
func NewRESTFactory(metaConnection resources.ArrayMetadata) resources.CollectorFactory {
	return &restFactory{metaConnection: metaConnection}
}

func (r *restFactory) InitializeCollector(arrayInfo *resources.ArrayRegistrationInfo) (resources.ArrayCollector, error) {
	switch arrayInfo.DeviceType {
	case common.FlashArray:
		return flasharray.NewCollector(arrayInfo.ID, arrayInfo.Name, arrayInfo.MgmtEndpoint, arrayInfo.APIToken, r.metaConnection)
	case common.FlashBlade:
		return flashblade.NewCollector(arrayInfo.ID, arrayInfo.Name, arrayInfo.MgmtEndpoint, arrayInfo.APIToken, r.metaConnection)
	default:
		return nil, fmt.Errorf("Unknown DeviceType")
	}
}
