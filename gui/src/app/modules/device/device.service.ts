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
 * Provides a service to fetch a list of storage devices.
 * Could fetch from a mocked backend or from an HTTP backend.
 */
import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { merge, Observable, of, Subject, timer } from 'rxjs';
import { catchError, delayWhen, map, repeatWhen, retryWhen, tap } from 'rxjs/operators';
import { Collection } from '../../shared/interfaces/collection';
import { DataPage } from '../../shared/interfaces/data-page';
import { ListParams, SortParams } from '../../shared/interfaces/list-params';
import { StorageDevice } from '../../shared/models/storage-device';

@Injectable({
  providedIn: 'root'
})
export class DeviceService implements Collection<StorageDevice> {
  // Should be triggered by calling code whenever a change is made that could require refreshing data.
  // Static so all instances of this service get triggered.
  static notifyRefresh: Subject<void> = new Subject<void>();

  API_SERVER_ADDRESS = '/api';
  DEVICES_ENDPOINT = '/arrays';

  constructor(private http: HttpClient) { }

  create(properties: Partial<StorageDevice>): Observable<StorageDevice> {
    return this.http.post<StorageDevice[]>(this.API_SERVER_ADDRESS + this.DEVICES_ENDPOINT, properties).pipe(map((deviceArray) => {
      return deviceArray[0]; // Only get the first item
    })).pipe(tap(() => setTimeout(() => DeviceService.notifyRefresh.next(), 1000))); // Trigger a delayed refresh
  }

  delete(id: string): Observable<string> {
    return this.http.delete(this.API_SERVER_ADDRESS + this.DEVICES_ENDPOINT, { params: { ids: id } }).pipe(map(x => {
      return x as string;
    })).pipe(tap(() => setTimeout(() => DeviceService.notifyRefresh.next(), 1000))); // Trigger a delayed refresh
  }

  list(params?: ListParams): Observable<DataPage<StorageDevice>> {
    // Default to sorting name ascending (since it just makes the
    // most sense)
    const queryParams = {
      'sort': 'name'
    };
    if (params) {
      if (params.filter) {
        Object.keys(params.filter).forEach((key: string) => {
          queryParams[key] = `*${params.filter[key]}*`; // Make each one a wildcard
        });
      }
      if (params.sort && (<SortParams>params.sort).key && (<SortParams>params.sort).order) {
        const cast = <SortParams>params.sort;
        if (cast.order === 'desc') {
          queryParams['sort'] = params.sort['key'] + '-';
        } else {
          queryParams['sort'] = params.sort['key'];
        }
      }
    }

    return this.http.get<StorageDevice[]>(this.API_SERVER_ADDRESS + this.DEVICES_ENDPOINT,
      { params: queryParams }).pipe(repeatWhen(delayWhen(() => {
      // Repeatedly poke the device service
      console.debug('Device list fetch successful.');
      return merge(timer(5000), DeviceService.notifyRefresh.asObservable());
    })), retryWhen(delayWhen(() => {
      // Repeatedly poke the device service, slower since we're having a hard time talking to it.
      console.log('Device list fetch unsuccessful, retrying.');
      return merge(timer(10000), DeviceService.notifyRefresh.asObservable());
    }))).pipe(map(x => { // Convert to DataPage
      return { total: x['response'].length, response: x['response'] };
    })).pipe(catchError(_ => { // In case an error somehow slips through, worst case return an empty page.
      return of({ total: 0, response: []});
    }));
  }
}
