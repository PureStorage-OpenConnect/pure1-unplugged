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

/**
 * This component displays the cards for all the devices fetched from
 * the device service.
 */
import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { ListParams } from '../../../../shared/interfaces/list-params';
import { DeviceAlert } from '../../../../shared/models/device-alert';
import { DeviceAlertQuery } from '../../../../shared/models/device-alert-query';
import { StorageDevice } from '../../../../shared/models/storage-device';
import { DeviceTagsService } from '../../../../shared/services/device-tags.service';
import { DeviceAlertService } from '../../../device-alert/device-alert.service';
import { DeviceService } from '../../device.service';
import { DeviceRegistrationModalComponent } from '../device-registration-modal/device-registration-modal.component';

@Component({
  selector: 'device-list',
  templateUrl: './device-list.component.html',
  styleUrls: ['./device-list.component.scss']
})
export class DeviceListComponent implements OnInit, OnDestroy {
  devices: StorageDevice[] = [];
  tags = new Map<string, Map<string, Map<string, string>>>();
  openAlerts = new Map<string, DeviceAlert[]>();
  destroy$: Subject<boolean>;
  deviceCancelSubject: Subject<void>;

  nameQuery: string;
  modelQuery: string;
  versionQuery: string;
  sortQuery: string;
  sortDir: 'asc' | 'desc';

  filter: any[];

  constructor(private deviceService: DeviceService, private tagsService: DeviceTagsService, private alertService: DeviceAlertService,
              private modalService: NgbModal, private router: Router, private route: ActivatedRoute) { }

  ngOnInit() {
    this.destroy$ = new Subject<boolean>();
    this.tagsService.list().pipe(takeUntil(this.destroy$)).subscribe(page => {
      this.tags = page.response[0];
    });
    this.alertService.list(new DeviceAlertQuery().withStatus('open')).pipe(takeUntil(this.destroy$)).subscribe(alerts => {
      this.openAlerts.clear();
      alerts.results.forEach(alert => {
        if (!this.openAlerts.has(alert.ArrayID)) {
          this.openAlerts.set(alert.ArrayID, []);
        }

        this.openAlerts.get(alert.ArrayID).push(alert);
      });
    });
    this.route.queryParams.subscribe(params => {
      if (params['filter']) {
        this.filter = JSON.parse(params['filter']);
      } else {
        this.filter = [];
      }

      if (params['sort']) {
        this.parseSortString(params['sort']);
      } else {
        this.sortQuery = 'name';
        this.sortDir = 'asc';
      }

      // Set model variables for the text boxes
      this.filter.forEach(filter => {
        if (filter.entity !== 'device') { return; }

        if (filter.key === 'array') { this.nameQuery = filter.value; }
        if (filter.key === 'model') { this.modelQuery = filter.value; }
        if (filter.key === 'version') { this.versionQuery = filter.value; }
      });

      this.reconnectDeviceService();
    });
  }

  ngOnDestroy(): void {
    console.debug('Unsubscribing from services');
    this.destroy$.next(true);
    if (this.deviceCancelSubject) {
      this.deviceCancelSubject.next();
    }
  }

  registerButtonClick(): void {
    this.modalService.open(DeviceRegistrationModalComponent, { backdrop: 'static', keyboard: false }).result.then((output) => {
      this.deviceService.create(output).subscribe((createdArray) => {

      });
    }).catch((reason: any) => {
      // Do nothing, modal dismissed
    });
  }

  trackById(index, item) {
    return item.id;
  }

  // Renavigates to the devices page, applying the filter and sort
  renavigate(clear: boolean): void {
    // Clean up the filters before renavigating: while this doesn't ensure clean filters *all* the time (user-entered URLs could be dirty),
    // it at least makes sure that when we change them programmatically they end up clean.
    this.removeImproperFilters();
    const queryParams = {};
    if (this.filter.length > 0) {
      queryParams['filter'] = JSON.stringify(this.filter);
    }
    if (this.sortQuery && this.sortDir) {
      queryParams['sort'] = this.getSortString();
    }
    this.router.navigate(['/dash/arrays'], { queryParams: queryParams});
    if (clear) {
      this.devices = [];
    }
    this.reconnectDeviceService();
  }

  reconnectDeviceService() {
    if (this.deviceCancelSubject) {
      this.deviceCancelSubject.next();
    }

    this.deviceCancelSubject = new Subject<void>();

    const query: ListParams = {};
    if (this.sortQuery && this.sortDir) {
      query.sort = {
        key: this.sortQuery,
        order: this.sortDir
      };
    }
    query.filter = {};
    if (this.nameQuery) {
      query.filter['names'] = this.nameQuery;
    }
    if (this.modelQuery) {
      query.filter['models'] = this.modelQuery;
    }
    if (this.versionQuery) {
      query.filter['versions'] = this.versionQuery;
    }

    this.deviceService.list(query).pipe(takeUntil(this.deviceCancelSubject)).subscribe(page => {
      this.devices = page.response;
    });
  }

  // Removes filters that are missing a field or have empty fields
  removeImproperFilters(): void {
    this.filter = this.filter.filter(f => f.hasOwnProperty('entity') && f.entity !== '' &&
                                     f.hasOwnProperty('key') && f.key !== '' &&
                                     f.hasOwnProperty('value') && f.value !== '');
  }

  getSortString(): string {
    if (this.sortDir === 'desc') {
      return this.sortQuery + '-';
    } else if (this.sortDir === 'asc') {
      return this.sortQuery;
    }
  }

  parseSortString(toParse: string): void {
    if (toParse.endsWith('-')) {
      this.sortQuery = toParse.substring(0, toParse.length - 1);
      this.sortDir = 'desc';
    } else {
      this.sortQuery = toParse;
      this.sortDir = 'asc';
    }
  }

  // Sets (creates/overwrites) the value for a key in the filter
  setFilterKeyValue(key: string, value: string): void {
    let keyExists = false;
    this.filter.forEach(filter => {
      if (filter.entity !== 'device' || filter.key !== key) { return; }

      filter.value = value;
      keyExists = true;
    });
    if (!keyExists) {
      this.filter.push({ entity: 'device', key: key, value: value });
    }
  }

  onDeviceNameInputChange(event): void {
    this.nameQuery = event;
    this.setFilterKeyValue('array', this.nameQuery);
    this.renavigate(false);
  }

  onDeviceModelInputChange(event): void {
    this.modelQuery = event;
    this.setFilterKeyValue('model', this.modelQuery);
    this.renavigate(false);
  }

  onVersionInputChange(event): void {
    this.versionQuery = event;
    this.setFilterKeyValue('version', this.versionQuery);
    this.renavigate(false);
  }

  onSortChange(event): void {
    this.sortQuery = event;
    // Because we're tracking by ID to maintain order, we need to actually *clear* the whole list
    // here to force it to reorder. Unfortunate consequence is all cards will flip back over, but
    // that's minor
    this.renavigate(true);
  }

  onSortDirChange(event): void {
    this.sortDir = event;
    // See onSortChange comment
    this.renavigate(true);
  }

  clearButtonClick(): void {
    this.nameQuery = '';
    this.setFilterKeyValue('array', this.nameQuery);
    this.modelQuery = '';
    this.setFilterKeyValue('model', this.modelQuery);
    this.versionQuery = '';
    this.setFilterKeyValue('version', this.versionQuery);
    this.sortQuery = 'name';
    this.sortDir = 'asc';
    // See onSortChange comment
    this.renavigate(true);
  }
}
