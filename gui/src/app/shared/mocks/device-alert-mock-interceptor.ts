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
export class DeviceAlertMockInterceptor implements HttpInterceptor {
    static ALERTS: Map<string, any[]>;
    static ALERT_TYPES = [
        {
            'component_name': 'c75-4',
            'event': 'failure',
            'severity': 'warning',
            'code': 46
        },
        {
            'component_name': 'ct0',
            'event': 'failure',
            'severity': 'critical',
            'code': 4
        },
        {
            'component_name': 'ct1',
            'event': 'failure',
            'severity': 'critical',
            'code': 4
        },
        {
            'component_name': 'infra',
            'event': 'delayed',
            'severity': 'warning',
            'code': 51
        },
        {
            'component_name': 'array.directory_service',
            'event': 'pureds test failure',
            'severity': 'warning',
            'code': 30
        },
        {
            'component_name': 'ct1.fc5',
            'event': 'Fibre Channel link failure',
            'severity': 'warning',
            'code': 39
        },
        {
            'component_name': 'ct0.fc3',
            'event': 'Fibre Channel link failure',
            'severity': 'warning',
            'code': 39
        }
    ];

    constructor() {
        DeviceMockUtil.initialize();
        DeviceAlertMockInterceptor.ALERTS = new Map<string, any>();
        for (let i = 0; i < DeviceMockUtil.fakeIds.length; i++) {
            const alerts = [];
            const alertsCount = Math.round(Math.random() * 4);
            for (let j = 0; j < alertsCount; j++) {
                alerts.push(DeviceAlertMockInterceptor.generateFakeAlert(DeviceMockUtil.fakeIds[i]));
            }
            DeviceAlertMockInterceptor.ALERTS.set(DeviceMockUtil.fakeIds[i], alerts);
        }
    }

    static generateFakeAlert(id: string): any {
        const alertType = DeviceAlertMockInterceptor.ALERT_TYPES[Math.round(Math.random() * DeviceAlertMockInterceptor.ALERT_TYPES.length)];
        return {
            'severity': alertType.severity,
            'actual': null,
            'variables': null,
            'code': alertType.code,
            'knowledge_base_url': '',
            'created': 0,
            'notified': 0,
            'subject': '',
            'component_name': alertType.component_name,
            'expected': null,
            'description': '',
            'index': 0,
            'opened': 0,
            'component_type': '',
            'component': '',
            'flagged': false,
            'name': '',
            'action': '',
            'details': '',
            'id': 0,
            'state': 'open',
            'category': 'array',
            'event': alertType.event,
            'updated': 0
        };
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        if (req.url === '/elasticsearch/pure1-unplugged/metrics/_search') {
            switch (req.method) {
                case 'POST': {
                    const body = JSON.parse(req.body);
                    try {
                        if (body.aggs) {
                            // Aggregations are what define an alerts fetch: this has them, so it's probably alerts
                            const buckets = [];

                            DeviceMockUtil.fakeIds.forEach(id => {
                                buckets.push({
                                    'top_result': {
                                        'hits': {
                                            'hits': [
                                                {
                                                    '_source': {
                                                        'DeviceID': id,
                                                        'Hostname': '',
                                                        'Alerts': DeviceAlertMockInterceptor.ALERTS.get(id)
                                                    }
                                                }
                                            ]
                                        }
                                    }
                                });
                            });

                            return of(new HttpResponse<any>({status: 200, body: {
                                'aggregations': {
                                    'unique_id': {
                                        'buckets': buckets
                                    }
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
