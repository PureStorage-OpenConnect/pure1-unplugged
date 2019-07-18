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

import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { MainPageComponent } from '../core/components/main-page/main-page.component';
import { ArrayCapacityComponent } from '../dashboard/components/array-capacity/array-capacity.component';
import { ArrayPerformanceComponent } from '../dashboard/components/array-performance/array-performance.component';
import { CapacityDashboardComponent } from '../dashboard/components/capacity-dashboard/capacity-dashboard.component';
import {
  FilesystemPerformanceComponent
} from '../dashboard/components/filesystem-performance/filesystem-performance.component';
import { MainDashboardComponent } from '../dashboard/components/main-dashboard/main-dashboard.component';
import { PerformanceDashboardComponent } from '../dashboard/components/performance-dashboard/performance-dashboard.component';
import { VolumePerformanceComponent } from '../dashboard/components/volume-performance/volume-performance.component';
import { AlertsViewComponent } from '../device-alert/components/alerts-view/alerts-view.component';
import { DeviceAlertModule } from '../device-alert/device-alert.module';
import { DeviceListComponent } from '../device/components/device-list/device-list.component';
import { SupportComponent } from '../support/components/support/support.component';

const routes: Routes = [
  { path: '', redirectTo: 'dash', pathMatch: 'full'},
  { path: 'dash', component: MainPageComponent,
    children: [
      // Default path/page (what the user sees when they first log in) is set here in "redirectTo"
      { path: '', redirectTo: 'arrays', pathMatch: 'full'},
      { path: 'dashboard', component: MainDashboardComponent },
      { path: 'arrays', component: DeviceListComponent },
      { path: 'analytics/performance', component: PerformanceDashboardComponent,
        children: [
          { path: 'arrays', component: ArrayPerformanceComponent },
          { path: 'filesystems', component: FilesystemPerformanceComponent },
          { path: 'volumes', component: VolumePerformanceComponent },
          { path: '**', redirectTo: 'arrays' }
        ]
      },
      { path: 'analytics/capacity', component: CapacityDashboardComponent,
        children: [
          { path: 'arrays', component: ArrayCapacityComponent },
          { path: '**', redirectTo: 'arrays' }
        ]
      },
      { path: 'analytics', redirectTo: 'analytics/performance' }, // Default analytics path
      { path: 'troubleshooting', component: SupportComponent },
      { path: 'messages', component: AlertsViewComponent }
    ]
  },
  { path: '**', redirectTo: 'dash'}
];

@NgModule({
  exports: [ RouterModule ],
  imports: [ RouterModule.forRoot(routes), DeviceAlertModule ]
})
export class AppRoutingModule {}
