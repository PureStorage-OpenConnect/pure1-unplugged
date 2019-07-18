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

import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { InlineSVGModule } from 'ng-inline-svg';
import { CoreModule } from '../core/core.module';
import { ArrayCapacityComponent } from './components/array-capacity/array-capacity.component';
import { ArrayPerformanceComponent } from './components/array-performance/array-performance.component';
import { CapacityDashboardComponent } from './components/capacity-dashboard/capacity-dashboard.component';
import { FilesystemPerformanceComponent } from './components/filesystem-performance/filesystem-performance.component';
import { MainDashboardComponent } from './components/main-dashboard/main-dashboard.component';
import { PerformanceDashboardComponent } from './components/performance-dashboard/performance-dashboard.component';
import { VolumePerformanceComponent } from './components/volume-performance/volume-performance.component';

@NgModule({
  imports: [
    CommonModule,
    RouterModule,
    CoreModule,
    InlineSVGModule.forRoot(),
    FormsModule
  ],
  declarations: [
    ArrayCapacityComponent,
    ArrayPerformanceComponent,
    CapacityDashboardComponent,
    FilesystemPerformanceComponent,
    MainDashboardComponent,
    PerformanceDashboardComponent,
    VolumePerformanceComponent
  ]
})
export class DashboardModule { }
