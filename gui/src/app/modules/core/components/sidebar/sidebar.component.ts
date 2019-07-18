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

import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import * as moment from 'moment';
import { filter } from 'rxjs/operators';
import { environment } from '../../../../../environments/environment';
import { ISidebarTab } from '../../../../shared/interfaces/sidebar-tab';
import { DeviceService } from '../../../device/device.service';

@Component({
  selector: 'sidebar',
  templateUrl: './sidebar.component.html',
  styleUrls: ['./sidebar.component.scss'],
  providers: [ RouterModule ]
})
export class SidebarComponent implements OnInit {
  static tabs: ISidebarTab[] = [{
    title: 'Dashboard',
    active: false,
    link: '/dashboard',
    icon: 'assets/images/nginclude/sidenav-dashboard.svg'
  }, {
    title: 'Arrays',
    active: false,
    link: '/arrays',
    icon: 'assets/images/nginclude/sidenav-array.svg'
  }, {
    title: 'Analytics',
    active: false,
    link: '/analytics',
    icon: 'assets/images/nginclude/sidenav-analytics.svg',
    subTabs: [{
      title: 'Performance',
      active: false,
      link: '/analytics/performance'
    }, {
      title: 'Capacity',
      active: false,
      link: '/analytics/capacity'
    }]
  }, {
    title: 'Messages',
    active: false,
    link: '/messages',
    icon: 'assets/images/nginclude/sidenav-message.svg'
  }, {
    title: 'Troubleshooting',
    active: false,
    link: '/troubleshooting',
    icon: 'assets/images/nginclude/sidenav-support.svg'
  }];

  connectedCount = 1;
  totalCount = 1;

  retracted = false;
  @Input() smallWindow = false;
  @Output() toggled: EventEmitter<boolean> = new EventEmitter<boolean>();
  tabs: ISidebarTab[] = [];
  activeTab: ISidebarTab = null;
  appVersion = '';
  apiLastSeen: moment.Moment = null;

  constructor(private router: Router, private deviceService: DeviceService) {
    this.tabs = SidebarComponent.tabs;
    this.appVersion = environment.version;
  }

  ngOnInit() {
    this.updateTabActiveStates(); // Call once the first time since the router doesn't seem to fire this event
    this.router.events.pipe(filter(event => event instanceof NavigationEnd)).subscribe(val => {
      this.updateTabActiveStates();
    });
    this.deviceService.list().subscribe(devices => {
      this.totalCount = devices.response.length;
      this.connectedCount = devices.response.reduce((count, device) => {
        return count + (device.status === 'Connected' ? 1 : 0);
      }, 0);
      this.apiLastSeen = moment();
    });
  }

  isTabActive(tab: ISidebarTab): boolean {
    return this.router.url.startsWith('/dash' + tab.link); // Comparing "startsWith" for the sake of matching subtabs
  }

  updateTabActiveStates(): void {
    this.tabs.forEach((tab: ISidebarTab) => {
      tab.active = this.isTabActive(tab);
      if (tab.active) { // This should always be only one tab, if it isn't something is dreadfully wrong
        this.activeTab = tab;
      }
      if (tab.subTabs) {
        tab.subTabs.forEach((subtab: ISidebarTab) => {
          subtab.active = this.isTabActive(subtab);
          if (subtab.active) {
            this.activeTab = subtab;
          }
        });
      }
    });
  }

  onSidebarToggleClick(): void {
    this.retracted = !this.retracted;
    this.toggled.emit(this.retracted);
  }

  menuClick(): void {
    // If the window is in the smaller state, make the sidebar retract again
    if (this.smallWindow) {
      this.retracted = true;
    }
  }

  getServerStatusString(): string {
    if (this.apiLastSeen == null) {
      return 'trying to connect';
    }
    const diff = moment.duration(moment().diff(this.apiLastSeen, 'seconds'), 'seconds');
    if (diff.seconds() > 20) { // ~2 error periods worth
      return 'disconnected';
    }
    return 'connected';
  }
}
