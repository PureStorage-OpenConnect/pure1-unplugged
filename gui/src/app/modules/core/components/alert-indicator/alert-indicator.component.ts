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

import { Component, Input } from '@angular/core';

type severity = 'critical' | 'warning' | 'info' | 'none';

@Component({
  selector: 'alert-indicator',
  templateUrl: './alert-indicator.component.html'
})
export class AlertIndicatorComponent {
  @Input() severity: severity;
  @Input() isDisabled = false;
  @Input() iconWidth = '20px';
  @Input() iconHeight = '20px';

  data = {
      none:     { svg: 'assets/icons/secondary/alert_info.svg',     class: 'pstg-card-flip-icon' },
      info:     { svg: 'assets/icons/secondary/alert_info.svg',     class: 'pstg-alert-info-icon' },
      warning:  { svg: 'assets/icons/secondary/alert_warning.svg',  class: 'pstg-alert-warning-icon' },
      critical: { svg: 'assets/icons/secondary/alert_critical.svg', class: 'pstg-alert-critical-icon' }
  };

  constructor() { }
}
