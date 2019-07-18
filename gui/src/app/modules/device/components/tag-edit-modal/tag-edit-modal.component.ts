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

import { Component, Input, OnInit } from '@angular/core';
import { FormControl } from '@angular/forms';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { DeviceTag } from '../../../../shared/models/device-tag';
import { StorageDevice } from '../../../../shared/models/storage-device';

@Component({
  selector: 'tag-edit-modal',
  templateUrl: './tag-edit-modal.component.html',
  styleUrls: ['./tag-edit-modal.component.scss']
})
export class TagEditModalComponent implements OnInit {

  testKey = 'asdfasdf';
  testValue = 'fdsafdas';

  tagsText = new FormControl('');
  tags: DeviceTag[] = [];

  @Input() device: StorageDevice;

  constructor(private activeModal: NgbActiveModal) { }

  ngOnInit() {
  }

  containsUserEnteredTags(): boolean {
    return this.tags.filter(tag => tag.namespace === 'pure1-unplugged').length > 0;
  }

  deleteTagClick(index: number): void {
    this.tags.splice(index, 1);
  }

  addTagClick(): void {
    this.tags.push({key: '', value: '', namespace: 'pure1-unplugged'});
  }

  containsInvalidTag(): boolean {
    return this.tags.some(tag => tag.key.trim().length === 0 || tag.value.trim().length === 0);
  }

  // Used to maintain consistency in the tags list
  trackByIndex(index, item) {
    return index;
  }

  onCancelClick(): void {
    this.activeModal.dismiss();
  }

  onSubmitClick(): void {
    this.activeModal.close(this.tags);
  }
}
