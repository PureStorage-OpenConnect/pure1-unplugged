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

import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { NgbTooltip } from '@ng-bootstrap/ng-bootstrap';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { DeviceAlert } from '../../../../shared/models/device-alert';
import { DeviceAlertQuery } from '../../../../shared/models/device-alert-query';
import { PaginationIndicatorComponent } from '../../../core/components/pagination-indicator/pagination-indicator.component';
import { DeviceAlertService } from '../../device-alert.service';

@Component({
  selector: 'alerts-view',
  templateUrl: './alerts-view.component.html',
  styleUrls: ['./alerts-view.component.scss'],
  providers: [ NgbTooltip ]
})
export class AlertsViewComponent implements OnInit {
  @ViewChild(PaginationIndicatorComponent) pagination: PaginationIndicatorComponent;
  readonly PAGE_SIZE: number = 50; // Constant

  alerts: DeviceAlert[] = [];
  alertCancelSubject: Subject<void>;

  states: string[];

  filter: any[];
  sort: string;
  sortDesc: boolean;

  deviceQuery: string;
  summaryQuery: string;
  statusQuery: string;
  severityQuery: string;

  constructor(private alertService: DeviceAlertService, private route: ActivatedRoute, private router: Router) { }

  ngOnInit() {
    this.reconnectAlertService();

    this.route.queryParams.subscribe(params => {
      if (params['filter']) {
        this.filter = JSON.parse(params['filter']);
      } else {
        this.filter = [];
      }

      if (params['sort']) {
        this.sortDesc = params['sort'].endsWith('-');
        if (this.sortDesc) {
          this.sort = params['sort'].substring(0, params['sort'].length - 1);
        } else {
          this.sort = params['sort'];
        }
      } else {
        this.sort = '';
      }

      if (params['page']) {
        this.pagination.currentPage = parseInt(params['page'], 10);
      } else {
        this.pagination.currentPage = 0;
      }

      // Default to 'all' if there's no status parameter
      this.statusQuery = 'all';

      // Set model variables for the text boxes
      this.filter.forEach(filter => {
        if (filter.entity !== 'alerts') { return; }

        if (filter.key === 'array') { this.deviceQuery = filter.value; }
        if (filter.key === 'summary') { this.summaryQuery = filter.value; }
        if (filter.key === 'status') { this.statusQuery = filter.value; }
        if (filter.key === 'severity') { this.severityQuery = filter.value; }
      });

      this.reconnectAlertService();
    });
  }

  // Makes a DeviceAlertQuery object from the current parameters
  constructAlertQuery(): DeviceAlertQuery {
    const query = new DeviceAlertQuery();
    if (this.deviceQuery && this.deviceQuery !== '') {
      query.withDeviceName(this.deviceQuery.toLowerCase()); // Elastic is set up to search with lowercase
    }
    if (this.summaryQuery && this.summaryQuery !== '') {
      query.withAlertSummary(this.summaryQuery.toLowerCase()); // Elastic is set up to search with lowercase
    }
    // ignore "all", since that's basically the same as not having a query for it at all
    if (this.statusQuery && this.statusQuery !== '' && this.statusQuery !== 'all') {
      query.withStatus(this.statusQuery);
    }
    if (this.severityQuery && this.severityQuery !== '') {
      query.withSeverity(this.severityQuery);
    }
    query.withPageSize(this.PAGE_SIZE);
    query.withPage(this.pagination.currentPage);
    query.withSort(this.sort, this.sortDesc ? 'desc' : 'asc');

    return query;
  }

  // Stops the existing connection, then makes a new one
  reconnectAlertService() {
    if (this.alertCancelSubject) {
      this.alertCancelSubject.next();
    }

    this.alertCancelSubject = new Subject<void>();

    this.alertService.list(this.constructAlertQuery()).pipe(takeUntil(this.alertCancelSubject)).subscribe(alerts => {
      this.alerts = alerts.results;
      this.pagination.itemCount = alerts.totalHits;
      this.states = alerts.states;
    });
  }

  // Renavigates to the messages page, applying the filter, sort, and current page
  renavigate(clear: boolean): void {
    // Clean up the filters before renavigating: while this doesn't ensure clean filters *all* the time (user-entered URLs could be dirty),
    // it at least makes sure that when we change them programmatically they end up clean.
    this.removeImproperFilters();
    const queryParams = {};
    if (this.filter.length > 0) {
      queryParams['filter'] = JSON.stringify(this.filter);
    }
    if (this.sort) {
      queryParams['sort'] = this.sort;
      if (this.sortDesc) {
        queryParams['sort'] += '-';
      }
    }
    if (this.pagination.currentPage !== 0) {
      queryParams['page'] = this.pagination.currentPage;
    }
    const urlWithoutParams = this.router.parseUrl(this.router.url).root.children['primary'].segments.map(it => '/' + it.path).join('');
    this.router.navigate([urlWithoutParams], { queryParams: queryParams});
    if (clear) {
      this.alerts = [];
    }
  }

  // Sets (creates/overwrites) the value for a key in the filter
  setFilterKeyValue(key: string, value: string): void {
    let keyExists = false;
    this.filter.forEach(filter => {
      if (filter.entity !== 'alerts' || filter.key !== key) { return; }

      filter.value = value;
      keyExists = true;
    });
    if (!keyExists) {
      this.filter.push({ entity: 'alerts', key: key, value: value });
    }
  }

  // Removes filters that are missing a field or have empty fields
  removeImproperFilters(): void {
    this.filter = this.filter.filter(f => f.hasOwnProperty('entity') && f.entity !== '' &&
                                     f.hasOwnProperty('key') && f.key !== '' &&
                                     f.hasOwnProperty('value') && f.value !== '');
  }

  pageClick(direction: number): void {
    this.pagination.currentPage += direction;
    if (this.pagination.currentPage < 0) { this.pagination.currentPage = 0; }

    this.renavigate(true);
  }

  sortClick(param: string): void {
    // If we're already sorting this one, invert it
    if (this.sort === param) {
      this.sortDesc = !this.sortDesc;
    } else {
      this.sort = param;
      this.sortDesc = false;
    }
    this.renavigate(true);
  }

  onDeviceNameInputChange(event): void {
    this.deviceQuery = event;
    this.setFilterKeyValue('array', this.deviceQuery);
    this.pagination.currentPage = 0;
    this.renavigate(false);
  }

  onSummaryInputChange(event): void {
    this.summaryQuery = event;
    this.setFilterKeyValue('summary', this.summaryQuery);
    this.pagination.currentPage = 0;
    this.renavigate(false);
  }

  onStatusInputChange(event): void {
    this.statusQuery = event;
    if (this.statusQuery === 'all') {
      // Remove the status query altogether
      this.filter = this.filter.filter(filter => !(filter.entity === 'alerts' && filter.key === 'status'));
    } else {
      this.setFilterKeyValue('status', this.statusQuery);
    }
    this.pagination.currentPage = 0;
    this.renavigate(false);
  }

  getMaxPageCount(): number {
    return Math.floor(this.pagination.itemCount / this.PAGE_SIZE);
  }

  getCurrentPageEndingIndex(): number {
    if (this.pagination.currentPage >= this.getMaxPageCount()) {
      return this.pagination.itemCount;
    } else {
      return (this.pagination.currentPage + 1) * this.PAGE_SIZE;
    }
  }

  // Gets the style to apply to a column header given the parameter key
  getSortStyle(param: string): string {
    if (!this.sort) { return 'st-sort-none'; }
    if (this.sort !== param) { return 'st-sort-none'; }

    // This is the parameter being sorted, now which direction?
    return this.sortDesc ? 'st-sort-descent' : 'st-sort-ascent';
  }

  trackByState(index, state: string) {
    return state;
  }

  trackByAlertAndDeviceID(index, alert: DeviceAlert) {
    return alert.ArrayID + alert.AlertID;
  }
}
