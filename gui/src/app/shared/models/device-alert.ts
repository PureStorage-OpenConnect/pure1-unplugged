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

import * as moment from 'moment';

// Classes
type AlertStatusName = 'open' | 'closed';
type AlertSeverityName = 'info' | 'warning' | 'critical';

export class DeviceAlert {
    AlertID: number;
    ArrayDisplayName: string;
    ArrayHostname: string;
    ArrayID: string;
    ArrayName: string;
    Code: number;
    Created: moment.Moment;
    Severity: AlertSeverityName;
    SeverityIndex: number;
    State: AlertStatusName;
    Summary: string;

    /** Gets the Knowledge Base link for this alert */
    getKBLink(): string {
        return `https://pure1.purestorage.com/external/knowledge/?cid=Alert_${this.Code.toString().padStart(4, '0')}`;
    }
}
