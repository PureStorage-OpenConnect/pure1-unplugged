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
import { InlineSVGModule } from 'ng-inline-svg';
import { AlertIndicatorComponent } from './components/alert-indicator/alert-indicator.component';
import { DialComponent } from './components/dial/dial.component';
import { PaginationIndicatorComponent } from './components/pagination-indicator/pagination-indicator.component';

@NgModule({
    declarations: [
      AlertIndicatorComponent,
      DialComponent,
      PaginationIndicatorComponent,
    ],
    imports: [
      InlineSVGModule.forRoot(),
      CommonModule
    ],
    exports: [
      AlertIndicatorComponent,
      DialComponent,
      PaginationIndicatorComponent
    ],
    bootstrap: []
  })
  export class CoreModule { }
