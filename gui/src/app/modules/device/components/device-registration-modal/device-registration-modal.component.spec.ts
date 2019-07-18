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

import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ReactiveFormsModule } from '@angular/forms';
import { NgbActiveModal, NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { InlineSVGModule } from 'ng-inline-svg';
import { CoreModule } from '../../../core/core.module';
import { DeviceRegistrationModalComponent } from './device-registration-modal.component';

describe('DeviceRegistrationModalComponent', () => {
    // Shared test components
    // All should be set in beforeEach()
    let fixture: ComponentFixture<DeviceRegistrationModalComponent> = null;
    let component: DeviceRegistrationModalComponent = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [
                DeviceRegistrationModalComponent
            ],
            imports: [
                InlineSVGModule.forRoot(),
                CoreModule,
                ReactiveFormsModule,
                NgbModule.forRoot()
            ],
            providers: [
                NgbActiveModal
            ]
        }).compileComponents();
    }));
    beforeEach(() => {
        fixture = TestBed.createComponent(DeviceRegistrationModalComponent);
        component = fixture.debugElement.componentInstance;
        component.ngOnInit();
    });

    it('should create the component', () => {
        expect(component).toBeTruthy();
    });

    it('should have empty defaults', () => {
        const form = component.deviceForm;
        expect(form).toBeTruthy();
        expect(form.value.array_name).toBe('');
        expect(form.value.mgmt_endpoint).toBe('');
        expect(form.value.api_token).toBe('');
        expect(form.value.device_type).toBe('');
    });

    it('should update the device type when the api token is changed if we haven\'t changed the device type', async () => {
        // Entering a FlashBlade token should change the product type to FlashBlade
        component.deviceForm.patchValue({ api_token: 'T-someflashbladetoken' });
        expect(component.deviceForm.value.device_type).toBe('FlashBlade');

        // Entering a FlashArray token should change the product type to FlashArray
        component.deviceForm.patchValue({ api_token: 'someflasharraytoken' });
        expect(component.deviceForm.value.device_type).toBe('FlashArray');
    });

    it('shouldn\'t update the device type when the api token is changed if we\'ve already changed the device type', async () => {
        // Mark the component as dirty (like the user had entered input)
        component.deviceForm.controls['device_type'].markAsDirty();

        // Entering a FlashBlade token shouldn't change it
        component.deviceForm.patchValue({ api_token: 'T-someflashbladetoken' });
        expect(component.deviceForm.value.device_type).toBe('');

        // Nor should entering a FlashArray token
        component.deviceForm.patchValue({ api_token: 'someflasharraytoken' });
        expect(component.deviceForm.value.device_type).toBe('');
    });
});
