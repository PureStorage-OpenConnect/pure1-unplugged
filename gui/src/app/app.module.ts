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

import { HttpClientModule } from '@angular/common/http';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';
import { NgbTooltipModule } from '@ng-bootstrap/ng-bootstrap';
import { InlineSVGModule } from 'ng-inline-svg';
import { AppComponent } from './app.component';
import { AppRoutingModule } from './modules/app-routing/app-routing.module';
import { MainPageComponent } from './modules/core/components/main-page/main-page.component';
import { SidebarComponent } from './modules/core/components/sidebar/sidebar.component';
import { CoreModule } from './modules/core/core.module';
import { DashboardModule } from './modules/dashboard/dashboard.module';
import { DeviceModule } from './modules/device/device.module';
import { SupportModule } from './modules/support/support.module';

@NgModule({
  declarations: [
    AppComponent,
    MainPageComponent,
    SidebarComponent
  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    DeviceModule,
    DashboardModule,
    SupportModule,
    CoreModule,
    AppRoutingModule,
    InlineSVGModule.forRoot(),
    NgbTooltipModule.forRoot(),
    FormsModule,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
