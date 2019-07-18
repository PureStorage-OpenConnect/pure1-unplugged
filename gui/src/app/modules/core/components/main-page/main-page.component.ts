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

import { HttpClient } from '@angular/common/http';
import { Component, HostListener, OnInit, ViewChild } from '@angular/core';
import * as moment from 'moment';
import { CookieService } from 'ngx-cookie-service';
import { isNullOrUndefined } from 'util';
import { DeviceAlert } from '../../../../shared/models/device-alert';
import { DeviceAlertQuery } from '../../../../shared/models/device-alert-query';
import { DeviceAlertService } from '../../../device-alert/device-alert.service';
import { DeviceService } from '../../../device/device.service';
import { SidebarComponent } from '../sidebar/sidebar.component';

@Component({
  selector: 'main-page',
  templateUrl: './main-page.component.html',
  styleUrls: ['./main-page.component.scss']
})
export class MainPageComponent implements OnInit {
  static readonly invalidTokenCookieMessage = '[token cookie invalid]';

  openFilterItem = {
    'entity': 'alerts',
    'key': 'status',
    'value': 'open'
  };

  warningIconFilter = JSON.stringify([
    {
      'entity': 'alerts',
      'key': 'severity',
      'value': 'warning'
    },
    this.openFilterItem
  ]);

  criticalIconFilter = JSON.stringify([
    {
      'entity': 'alerts',
      'key': 'severity',
      'value': 'critical'
    },
    this.openFilterItem
  ]);

  @ViewChild(SidebarComponent)
  sidebar: SidebarComponent;
  retracted = false;
  logoutModalOpen = false;
  username: string;

  warningCount = 0;
  criticalCount = 0;
  registeredDeviceIds = [];

  smallWindow = false; // Whether the window is too small that we need to shade + shrink the sidebar all the way

  constructor(private cookieService: CookieService, private deviceService: DeviceService,
              private alertService: DeviceAlertService, private httpClient: HttpClient) { }

  static getTokenEmail(token: string): string {
    if (isNullOrUndefined(token)) {
      return MainPageComponent.invalidTokenCookieMessage;
    }

    const jwtSplit = token.split('.');
    if (jwtSplit.length !== 3) {
      return MainPageComponent.invalidTokenCookieMessage;
    }

    try {
      const claims = atob(jwtSplit[1]);
      const parsedClaims = JSON.parse(claims);
      if (!parsedClaims.hasOwnProperty('email')) {
        return MainPageComponent.invalidTokenCookieMessage;
      }
      return parsedClaims.email;
    } catch {
      return MainPageComponent.invalidTokenCookieMessage;
    }
  }

  static getTokenExpiryTime(token: string): number {
    if (isNullOrUndefined(token)) {
      return null;
    }

    const jwtSplit = token.split('.');
    if (jwtSplit.length !== 3) {
      return null;
    }

    try {
      const claims = atob(jwtSplit[1]);
      const parsedClaims = JSON.parse(claims);
      if (!parsedClaims.hasOwnProperty('exp')) {
        console.error('Couldn\'t find token expiration property');
        return null;
      }
      const expiresAt = parseInt(parsedClaims.exp, 10);
      if (isNaN(expiresAt)) {
        console.error('Token expiration time is not a number');
        return null;
      }
      return expiresAt - moment().unix();
    } catch {
      return null;
    }
  }

  ngOnInit() {
    if (this.cookieService.check('pure1-unplugged-token')) {
      const token = this.cookieService.get('pure1-unplugged-token');
      this.username = MainPageComponent.getTokenEmail(token);
      const expiresIn = MainPageComponent.getTokenExpiryTime(token);
      if (!isNullOrUndefined(expiresIn)) {
        console.debug('Token expires in ' + expiresIn + ' seconds');
        // If it's already expired, don't do anything (dunno how we got here in the first place anyways?)
        if (expiresIn > 0) {
          console.debug('Triggering expiration in ' + expiresIn + ' seconds');
          setTimeout(() => {
            window.alert('Your session has expired. Please log in again.');
            window.location.reload();
          }, (expiresIn + 5) * 1000); // Reload a few seconds after it expires, just to make really sure it's expired
        }
      }
    } else {
      this.username = '[no token cookie]';
    }

    this.deviceService.list().subscribe(page => {
      this.registeredDeviceIds = [];
      page.response.forEach(device => {
        this.registeredDeviceIds.push(device.id);
      });
    });

    this.alertService.list(new DeviceAlertQuery().withStatus('open')).subscribe(alerts => {
      const deviceIDs = this.registeredDeviceIds;
      this.warningCount = alerts.results.reduce(function(count: number, alert: DeviceAlert) {
        // Only count still-registered devices (this just updates the icon a bit faster)
        if (!deviceIDs.some(id => id === alert.ArrayID)) {
          return count;
        }
        return count + (alert.Severity === 'warning' && alert.State === 'open' ? 1 : 0);
      }, 0);
      this.criticalCount = alerts.results.reduce(function(count: number, alert: DeviceAlert) {
        // Only count still-registered devices (this just updates the icon a bit faster)
        if (!deviceIDs.some(id => id === alert.ArrayID)) {
          return count;
        }
        return count + (alert.Severity === 'critical' && alert.State === 'open' ? 1 : 0);
      }, 0);
    });
  }

  onSidebarToggleClick() {
    this.sidebar.onSidebarToggleClick();
  }

  onSidebarToggled(retracted: boolean): void {
    this.retracted = retracted;
  }

  onUserClick() {
    if (this.smallWindow && !this.retracted) { return; } // If the button is hidden behind the screen dimming, don't let it activate

    this.logoutModalOpen = !this.logoutModalOpen;
  }

  pageClick(event) {
    // If this isn't inside the user modal or the user display
    if (event.target.closest('.logout-modal-wrapper, .topbar-item-user') === null) {
      this.logoutModalOpen = false; // Close the logout modal
    }
  }

  @HostListener('window:resize', ['$event'])
  onWindowResize(event) {
    this.smallWindow = event.target.innerWidth < 768;
  }

  logoutClick() {
    const session = this.cookieService.get('pure1-unplugged-session');
    this.httpClient.delete('/auth/api-token', {
      params: {
        'name': session
      },
      responseType: 'text'
    }).subscribe(null, null, () => {
      this.cookieService.delete('pure1-unplugged-token', '/');
      this.cookieService.delete('pure1-unplugged-session', '/');
      window.location.reload();
    }); // Only handle success, not on error
  }
}
