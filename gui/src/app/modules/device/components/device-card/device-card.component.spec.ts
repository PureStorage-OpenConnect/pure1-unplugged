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
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { InlineSVGModule } from 'ng-inline-svg';
import { NgSlimScrollModule } from 'ngx-slimscroll';
import { CoreModule } from '../../../core/core.module';
import { DeviceCardComponent } from './device-card.component';

describe('DeviceCardComponent', () => {
    // Shared test components
    // All should be set in beforeEach()
    let fixture: ComponentFixture<DeviceCardComponent> = null;
    let component: any = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [
                DeviceCardComponent,
            ],
            imports: [
                InlineSVGModule.forRoot(),
                HttpClientModule,
                NgbModule.forRoot(),
                CoreModule,
                NgSlimScrollModule
            ],
        }).compileComponents();
    }));
    beforeEach(() => {
        fixture = TestBed.createComponent(DeviceCardComponent);
        component = fixture.debugElement.componentInstance;
    });

    it('should create the component', () => {
        expect(component).toBeTruthy();
    });

    describe('PercentageDisplay', () => {
        it('should process low decimals properly', async () => {
            expect(DeviceCardComponent.getRoundedPercentage(0.05123)).toBe(5);
        });
        it('should process middle decimals properly', async () => {
            expect(DeviceCardComponent.getRoundedPercentage(0.055)).toBe(5);
        });
        it('should process high decimals properly', async () => {
            expect(DeviceCardComponent.getRoundedPercentage(0.05987)).toBe(5);
        });
    });
});
