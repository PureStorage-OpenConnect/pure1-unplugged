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

import { HttpEvent, HttpHandler, HttpInterceptor, HttpRequest, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { DeviceTag } from '../models/device-tag';
import { DeviceMockUtil } from './device-mock-util';

@Injectable()
export class DeviceTagMockInterceptor implements HttpInterceptor {
    static TAGS: Map<string, DeviceTag[]>;

    constructor() {
        DeviceMockUtil.initialize();
        DeviceTagMockInterceptor.TAGS = new Map<string, DeviceTag[]>();
        DeviceMockUtil.fakeIds.forEach(id => {
            DeviceTagMockInterceptor.TAGS.set(id, DeviceTagMockInterceptor.generateDefaultTags());
        });
    }

    static generateDefaultTags(): DeviceTag[] {
        return [
            {
                'key': 'purestorage.com/backend',
                'namespace': 'psoNamespace',
                'value': 'block'
            },
            {
                'key': 'access-mode.purestorage.com/RWX',
                'namespace': 'psoNamespace',
                'value': 'false'
            },
            {
                'key': 'protection.purestorage.com/snapshot',
                'namespace': 'psoNamespace',
                'value': 'true'
            },
            {
                'key': 'purestorage.com/hostname',
                'namespace': 'psoNamespace',
                'value': ''
            },
            {
                'key': 'purestorage.com/family',
                'namespace': 'psoNamespace',
                'value': ''
            },
            {
                'key': 'access-mode.purestorage.com/RWO',
                'namespace': 'psoNamespace',
                'value': 'true'
            },
            {
                'key': 'access-mode.purestorage.com/ROX',
                'namespace': 'psoNamespace',
                'value': 'false'
            },
            {
                'key': 'purestorage.com/id',
                'namespace': 'psoNamespace',
                'value': '01234567-890a-bcde-f012-3456789abc'
            }
        ];
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        if (req.url === '/api/arrays/tags') {
            switch (req.method) {
                case 'GET': {
                    const toReturn = [];
                    if (req.params.has('ids')) {
                        toReturn.push({
                            'id': req.params.get('ids'),
                            'tags': DeviceTagMockInterceptor.TAGS.get(req.params.get('ids'))
                        });
                    } else {
                        DeviceMockUtil.fakeIds.forEach(id => {
                            toReturn.push({
                                'id': id,
                                'tags': DeviceTagMockInterceptor.TAGS.get(id)
                            });
                        });
                    }

                    return of(new HttpResponse<any>({ status: 200, body: toReturn}));
                }
                case 'PATCH': {
                    // Don't actually patch it, but pretend we did
                    return of(new HttpResponse<any>({ status: 200, body: [ 'Patch successful (mocked data)' ]}));
                }
                case 'DELETE': {
                    // Don't actually delete it, but pretend we did
                    return of(new HttpResponse<any>({ status: 200, body: 'Deletion successful (mocked data)' }));
                }
            }
        }

        return next.handle(req); // Return unmodified
    }
}
