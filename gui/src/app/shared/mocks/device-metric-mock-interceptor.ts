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
import { DeviceMockUtil } from './device-mock-util';

@Injectable()
export class DeviceMetricMockInterceptor implements HttpInterceptor {
    static METRICS: any[];

    constructor() {
        DeviceMockUtil.initialize();
        DeviceMetricMockInterceptor.METRICS = [];
        for (let i = 0; i < DeviceMockUtil.fakeIds.length; i++) {
            DeviceMetricMockInterceptor.METRICS.push(DeviceMetricMockInterceptor.generateFakeMetric(DeviceMockUtil.fakeIds[i]));
        }
    }

    static generateFakeMetric(id: string) {
        const volCount = Math.round(Math.random() * 2000);
        const activeVolCount = Math.round(Math.random() * volCount);
        const totalSpace = Math.random() * 9999999999999;
        const usedSpace = Math.random() * totalSpace;
        const percentFull = usedSpace / totalSpace;
        return {
            'ReadsPerSec': Math.round(Math.random() * 100),
            'HostCount': 71,
            'InputPerSec': Math.round(Math.random() * 100000000),
            'DeviceID': id,
            'OutputPerSec': Math.round(Math.random() * 100000000),
            'CreatedAt': 1538757340,
            'Hostname': '', // Blank at the moment as we don't have a nice way to get this
            'QueueDepth': 2,
            'UsedSpace': usedSpace,
            'PendingEradicateVolumeCount': 0,
            'WritesPerSec': Math.round(Math.random() * 100),
            'VolumeCount': volCount,
            'ActiveVolumeCount': activeVolCount,
            'SnapshotCount': Math.round(Math.random() * 100),
            'AlertMessageCount': 0,
            'HealthScore': 100,
            'TotalReduction': Math.random() * 200,
            'ReadLatency': Math.round(Math.random() * 2000),
            'PercentFull': percentFull,
            'TotalSpace': totalSpace,
            'DataReduction': (Math.random() * 6) + 1,
            'WriteLatency': Math.round(Math.random() * 2000)
        };
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        if (req.url === '/elasticsearch/pure1-unplugged/metrics/_search') {
            switch (req.method) {
                case 'POST': {
                    const body = JSON.parse(req.body);
                    try {
                        if (!body.aggs) {
                            // Aggregations are what define an alerts fetch: this has none, so it's probably metrics
                            const deviceID = body.query.match.DeviceID;

                            return of(new HttpResponse<any>({status: 200, body: {
                                'hits': {
                                    'hits': [
                                        {
                                            '_source': DeviceMetricMockInterceptor.METRICS.find(metric => metric.DeviceID === deviceID)
                                        }
                                    ]
                                }
                            }}));
                        }
                    } catch (error) {
                        return next.handle(req);
                    }
                }
            }
        }

        return next.handle(req); // Return unmodified
    }
}
