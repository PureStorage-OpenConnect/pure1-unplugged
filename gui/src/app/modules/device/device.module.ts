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
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { NgbModalModule, NgbTooltipModule } from '@ng-bootstrap/ng-bootstrap';
import { InlineSVGModule } from 'ng-inline-svg';
import { CookieService } from 'ngx-cookie-service';
import { NgSlimScrollModule } from 'ngx-slimscroll';
import { CoreModule } from '../core/core.module';
import { DeviceCardComponent } from './components/device-card/device-card.component';
import { DeviceDeleteModalComponent } from './components/device-delete-modal/device-delete-modal.component';
import { DeviceListComponent } from './components/device-list/device-list.component';
import { DeviceRegistrationCardComponent } from './components/device-registration-card/device-registration-card.component';
import { DeviceRegistrationModalComponent } from './components/device-registration-modal/device-registration-modal.component';
import { TagEditModalComponent } from './components/tag-edit-modal/tag-edit-modal.component';
import { TagListItemComponent } from './components/tag-edit-modal/tag-list-item/tag-list-item.component';

@NgModule({
    declarations: [
        DeviceCardComponent,
        DeviceDeleteModalComponent,
        DeviceListComponent,
        DeviceRegistrationCardComponent,
        DeviceRegistrationModalComponent,
        TagEditModalComponent,
        TagListItemComponent
    ],
    imports: [
        BrowserModule,
        InlineSVGModule.forRoot(),
        NgbModalModule.forRoot(),
        FormsModule,
        ReactiveFormsModule,
        NgbTooltipModule,
        CoreModule,
        CommonModule,
        RouterModule,
        NgSlimScrollModule
    ],
    exports: [
        DeviceCardComponent,
        DeviceListComponent
    ],
    entryComponents: [
        DeviceRegistrationModalComponent,
        TagEditModalComponent,
        DeviceDeleteModalComponent
    ],
    bootstrap: [],
    providers: [ CookieService ]
  })
  export class DeviceModule { }
