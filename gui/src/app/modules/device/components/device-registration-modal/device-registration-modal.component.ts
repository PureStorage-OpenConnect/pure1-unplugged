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

import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { StorageDevice } from '../../../../shared/models/storage-device';

@Component({
  selector: 'device-registration-modal',
  templateUrl: './device-registration-modal.component.html',
  styleUrls: ['./device-registration-modal.component.scss']
})
export class DeviceRegistrationModalComponent implements OnInit {
  deviceForm: FormGroup;

  constructor(private activeModal: NgbActiveModal, private fb: FormBuilder) {
    this.deviceForm = this.fb.group({
      array_name: ['', Validators.required ],
      mgmt_endpoint: ['', Validators.required ],
      api_token: ['', Validators.required ],
      device_type: ['', Validators.required ]
    });
    this.deviceForm.get('api_token').valueChanges.subscribe((value) => {
      // If the user already changed it, don't change it again!
      if (this.deviceForm.controls['device_type'].dirty) { return; }

      // Attempt to auto-fill the device type based on the API token
      if (value.startsWith('T-')) {
        this.deviceForm.patchValue({device_type: 'FlashBlade'}); // Auto-select FB based on the token, since it matches the pattern
      } else {
        this.deviceForm.patchValue({device_type: 'FlashArray'}); // Doesn't match the FB pattern, so go with FA as default
      }
    });
  }

  ngOnInit() { }

  onCancelClick(): void {
    this.activeModal.dismiss();
  }

  onSubmitClick(): void {
    const partial = this.deviceForm.value as Partial<StorageDevice>;
    partial.name = this.deviceForm.get('array_name').value;
    this.activeModal.close(partial);
  }
}
