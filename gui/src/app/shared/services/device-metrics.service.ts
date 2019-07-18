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
 * Provides a service to fetch the latest metrics for a device.
 */
import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, timer } from 'rxjs';
import { delayWhen, map, repeatWhen, retryWhen } from 'rxjs/operators';
import { DeviceMetric } from '../models/device-metric';

@Injectable({
  providedIn: 'root'
})
export class DeviceMetricsService {
    ELASTIC_ADDRESS = '/elasticsearch';
    METRIC_ENDPOINT = '/pure-arrays-metrics-*/_search';

    constructor(private http: HttpClient) { }

    getLatestMetric(deviceID: string): Observable<DeviceMetric> {
        return this.http.post(this.ELASTIC_ADDRESS + this.METRIC_ENDPOINT, `{
            "sort": { "CreatedAt": "desc" },
            "size": 1,
            "query": {
                "term": {
                    "ArrayID": {
                        "value": "${deviceID}"
                    }
                }
            },
            "_source": {
                "excludes": ["Tags"]
            }
        }`, { headers: { 'Content-Type': 'application/json'}})
                .pipe(map(result => result['hits']['hits'][0]['_source'])) // Get the part of the response we need
                .pipe(map(result => {
                    // Convert to the proper class
                    return Object.assign(new DeviceMetric(), result);
                }))
                .pipe(
                    repeatWhen(delayWhen(() => {
                        // Repeatedly poke the metrics service
                        console.debug(`Metric fetch for ${deviceID} successful.`);
                        return timer(5000);
                    })),
                    retryWhen(delayWhen(() => {
                        // Repeatedly poke the metrics service, slower since we're having a hard time talking to it.
                        console.log(`Metric fetch for ${deviceID} unsuccessful, retrying.`);
                        return timer(15000);
                    }))
                );
    }
}
