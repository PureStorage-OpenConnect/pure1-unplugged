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
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { NgbActiveModal, NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { DeviceDeleteModalComponent } from './device-delete-modal.component';

describe('DeviceDeleteModalComponent', () => {
  let component: DeviceDeleteModalComponent;
  let fixture: ComponentFixture<DeviceDeleteModalComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DeviceDeleteModalComponent ],
      imports: [
        FormsModule,
        ReactiveFormsModule,
        NgbModule.forRoot()
      ],
      providers: [
        NgbActiveModal
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DeviceDeleteModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('shouldn\'t allow deletion when the name doesn\'t match', async () => {
    component.deviceName = 'Some Device Name!'; // Set the name to test
    const submitButton = fixture.nativeElement.querySelector('input[type="submit"]');
    expect(submitButton.disabled).toBeTruthy();

    component.nameInput.patchValue('Some Other Name'); // Doesn't match
    fixture.detectChanges();
    await fixture.whenStable();
    expect(submitButton.disabled).toBeTruthy();

    component.nameInput.patchValue('some device name!'); // Equal but different case, doesn't match
    fixture.detectChanges();
    await fixture.whenStable();
    expect(submitButton.disabled).toBeTruthy();
  });

  it('should allow deletion when the name matches', async () => {
    component.deviceName = 'Some Device Name!'; // Set the name to test
    const submitButton = fixture.nativeElement.querySelector('input[type="submit"]');
    expect(submitButton.disabled).toBeTruthy();

    component.nameInput.patchValue('Some Device Name!');
    fixture.detectChanges();
    await fixture.whenStable();
    expect(submitButton.disabled).toBeFalsy();
  });
});
