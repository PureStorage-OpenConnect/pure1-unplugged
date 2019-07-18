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

import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import * as moment from 'moment';
import { Observable, timer } from 'rxjs';
import { delayWhen, map, repeatWhen, retryWhen } from 'rxjs/operators';
import { DeviceAlert } from '../../shared/models/device-alert';
import { DeviceAlertQuery } from '../../shared/models/device-alert-query';
import { DeviceAlertResult } from '../../shared/models/device-alert-result';

@Injectable({
  providedIn: 'root'
})
export class DeviceAlertService {
  ELASTIC_ADDRESS = '/elasticsearch';

  ALERTS_ENDPOINT = '/pure-alerts/alerts/_search';
  ALERTS_FILTER_PATH = 'hits.hits._source,hits.total,aggregations.all.states.buckets.key';

  constructor(private http: HttpClient) { }

  convertRawAlertToClass(rawAlert: any): DeviceAlert {
    const toReturn = new DeviceAlert();

    toReturn.AlertID = rawAlert['AlertID'];
    toReturn.Code = rawAlert['Code'];
    toReturn.Created = rawAlert['Created'] === 0 ? null : moment.unix(rawAlert['Created']);
    toReturn.ArrayHostname = rawAlert['ArrayHostname'];
    toReturn.ArrayID = rawAlert['ArrayID'];
    toReturn.ArrayName = rawAlert['ArrayName'];
    toReturn.ArrayDisplayName = rawAlert['ArrayDisplayName'];
    toReturn.Severity = rawAlert['Severity'];
    toReturn.SeverityIndex = rawAlert['SeverityIndex'];
    toReturn.State = rawAlert['State'];
    toReturn.Summary = rawAlert['Summary'];

    return toReturn;
  }

  list(query: DeviceAlertQuery = new DeviceAlertQuery()): Observable<DeviceAlertResult> {
    const requestObj = query.generateRequestObject();
    // Add an aggregation by state, so we can find whcih
    requestObj['aggs'] = {
      'all': {
        'global': {},
        'aggs': {
          'states': {
            'terms': {
              'field': 'State'
            }
          }
        }
      }
    };

    return this.http.post(this.ELASTIC_ADDRESS + this.ALERTS_ENDPOINT + '?filter_path=' + this.ALERTS_FILTER_PATH,
      JSON.stringify(requestObj),
      { headers: { 'Content-Type': 'application/json'}})
      .pipe(map(rawResponse => {
        // If we're missing fields, just set up empty default objects so that the rest of the flow works right
        if (!('hits' in rawResponse) ||
            !('hits' in rawResponse['hits'])) {
          rawResponse['hits']['hits'] = [];
        }
        if (!('aggregations' in rawResponse) ||
            !('all' in rawResponse['aggregations']) ||
            !('states' in rawResponse['aggregations']['all']) ||
            !('buckets' in rawResponse['aggregations']['all']['states'])) {
          rawResponse['aggregations'] = {
            'all': {
              'states': {
                'buckets': []
              }
            }
          };
        }

        return rawResponse;
      }))
      .pipe(map(rawHits => {
        const alerts = [];

        rawHits['hits']['hits'].forEach(hit => {
          alerts.push(this.convertRawAlertToClass(hit['_source']));
        });

        const states = [];
        rawHits['aggregations']['all']['states']['buckets'].forEach(bucket => {
          states.push(bucket['key']);
        });

        return new DeviceAlertResult(alerts, rawHits['hits']['total'], states);
      }))
      .pipe(repeatWhen(delayWhen(() => {
        // Repeatedly poke the alert service (but only every 15 seconds because we really don't need this to update often)
        console.debug('Alert list fetch successful.');
        return timer(15000);
      })), retryWhen(delayWhen(() => {
        // Repeatedly poke the alert service, slower since we're having a hard time talking to it.
        console.log('Alert list fetch unsuccessful, retrying.');
        return timer(30000);
      })));
  }
}
