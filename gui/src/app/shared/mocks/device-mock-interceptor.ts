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
import { StorageDevice } from '../models/storage-device';
import { DeviceMockUtil } from './device-mock-util';

@Injectable()
export class DeviceMockInterceptor implements HttpInterceptor {
    static ARRAYS: StorageDevice[] = [
        {
            id: '',
            name: 'I\'m An Array!',
            mgmt_endpoint: '192.168.1.1',
            api_token: 'AAAAAHHH',
            status: 'Connected',
            device_type: 'FlashArray',
            model: 'FA-420',
            version: '5.1.3',
            _as_of: '0',
            _last_updated: '0'
        }, {
            id: '',
            name: 'NOT AN ARRAY',
            mgmt_endpoint: '1.1.861.291',
            api_token: 'AAAAAHHH',
            status: 'Connected',
            device_type: 'FlashBlade',
            model: 'FlashBlade',
            version: '1.2.3',
            _as_of: '0',
            _last_updated: '0'
        }
    ];

    constructor() {
        DeviceMockUtil.initialize();
        for (let i = 0; i < DeviceMockUtil.fakeIds.length; i++) {
            DeviceMockInterceptor.ARRAYS[i].id = DeviceMockUtil.fakeIds[i];
        }
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        if (req.url === '/api/arrays') {
            switch (req.method) {
                case 'GET': {
                    if (req.params.has('ids')) {
                        return of(new HttpResponse<any>({ status: 200, body: DeviceMockInterceptor.ARRAYS
                            .filter(device => device.id === req.params.get('ids'))
                            .slice(0, DeviceMockInterceptor.ARRAYS.length) }));
                    }

                    return of(new HttpResponse<any>({ status: 200, body: DeviceMockInterceptor.ARRAYS
                        .slice(0, DeviceMockInterceptor.ARRAYS.length)}));
                }
                case 'POST': {
                    // Don't actually add it, but pretend we did
                    const newDevice = new StorageDevice();
                    newDevice.name = req.body.name;
                    newDevice.mgmt_endpoint = req.body.mgmt_endpoint;
                    newDevice.status = 'Connecting';
                    return of(new HttpResponse<any>({ status: 200, body: [ newDevice ]}));
                }
                case 'DELETE': {
                    // Don't actually delete it, but pretend we did
                    return of(new HttpResponse<any>({ status: 200, body: 'Deleted.' }));
                }
            }
        }

        return next.handle(req); // Return unmodified
    }
}
