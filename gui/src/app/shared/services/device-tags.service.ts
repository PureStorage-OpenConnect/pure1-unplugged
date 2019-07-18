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
 * Provides a service to fetch a list of tags for all arrays.
 * Could fetch from a mocked backend or from an HTTP backend.
 */
import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { merge, Observable, of, Subject, timer } from 'rxjs';
import { catchError, delayWhen, map, repeatWhen, retryWhen, tap } from 'rxjs/operators';
import { Collection } from '../interfaces/collection';
import { DataPage } from '../interfaces/data-page';
import { ListParams } from '../interfaces/list-params';

@Injectable({
  providedIn: 'root'
})
export class DeviceTagsService implements Collection<Map<string, Map<string, Map<string, string>>>> { // Device => namespace => key => value
  // Should be triggered by calling code whenever a change is made that could require refreshing data.
  // Static so all instances of this service get triggered.
  static notifyRefresh: Subject<void> = new Subject<void>();

  API_SERVER_ADDRESS = '/api';
  TAGS_ENDPOINT = '/arrays/tags';

  constructor(private http: HttpClient) { }

  create(properties: Partial<Map<string, Map<string, Map<string, string>>>>): Observable<Map<string, Map<string, Map<string, string>>>> {
    throw new Error('Method not implemented.');
  }

  delete(id: string): Observable<string> {
    throw new Error('Method not implemented.');
  }

  list(params?: ListParams): Observable<DataPage<Map<string, Map<string, Map<string, string>>>>> {
    console.debug('Beginning tag list fetch...');

    return this.http.get<any>(this.API_SERVER_ADDRESS + this.TAGS_ENDPOINT).pipe(repeatWhen(delayWhen(() => {
      // Repeatedly poke the tag service
      console.debug('Tag list fetch successful.');
      return merge(timer(5000), DeviceTagsService.notifyRefresh.asObservable());
    })), retryWhen(delayWhen(() => {
      // Repeatedly poke the tag service, slower since we're having a hard time talking to it.
      console.log('Tag list fetch unsuccessful, retrying.');
      return merge(timer(10000), DeviceTagsService.notifyRefresh.asObservable());
    }))).pipe(map((arrayIn) => {
      // Convert to a multi-layer map
      const toReturn = new Map<string, Map<string, Map<string, string>>>();
      arrayIn['response'].forEach((device) => {
        if (!toReturn.has(device.id)) { toReturn.set(device.id, new Map<string, Map<string, string>>()); }
        if (!device.tags) {
          // Usually this means that the array is disconnected
          toReturn.set(device.id, new Map<string, Map<string, string>>());
          return;
        }

        // Has 'id' and 'tags
        device.tags.forEach((tag) => {
          // Has 'key', 'namespace', and 'value'
          if (!toReturn.get(device.id).has(tag.namespace)) { toReturn.get(device.id).set(tag.namespace, new Map<string, string>()); }
          toReturn.get(device.id).get(tag.namespace).set(tag.key, tag.value);
        });
      });
      return toReturn;
    })).pipe(map(x => { // Convert to DataPage: encapsulate the map in an array to make it work.
      return { total: 1, response: [x] };
    })).pipe(catchError(err => { // In case an error somehow slips through, worst case return an empty page.
      console.error('Error fetching tags: ' + err);
      return of({ total: 0, response: []});
    }));
  }

  patch(id: string, tagPatches: {key: string, value: string}[]): Observable<any> {
    return this.http.patch(this.API_SERVER_ADDRESS + this.TAGS_ENDPOINT, { tags: tagPatches }, { params: { ids: id }})
          .pipe(tap(() => setTimeout(() => DeviceTagsService.notifyRefresh.next(), 1000))); // Trigger a delayed refresh
  }

  deleteTag(id: string, tag: string): Observable<any> {
    return this.http.delete(this.API_SERVER_ADDRESS + this.TAGS_ENDPOINT, { params: { ids: id, tags: tag } })
          .pipe(tap(() => setTimeout(() => DeviceTagsService.notifyRefresh.next(), 1000))); // Trigger a delayed refresh
  }
}
