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

// tslint:disable:interface-name
import { Observable } from 'rxjs';
import { DataPage } from './data-page';
import { ListParams } from './list-params';

export interface Collection<T> {
    create(properties: Partial<T>): Observable<T>;
    delete(id: string): Observable<string>;
    // get(name: string, array: string): Observable<T>;
    list(params?: ListParams): Observable<DataPage<T>>;
    // update(name: string, array: string, properties: {}): Observable<T>;
}
