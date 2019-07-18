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

export enum SortFieldEnum {
  device_name = 'ArrayDisplayName',
  severity = 'SeverityIndex',
  created = 'Created',
  summary = 'Summary.raw',
  state = 'State'
}
export type SortDirection = 'asc' | 'desc';

export class DeviceAlertQuery {
  deviceName = '';
  alertSummary = '';
  status = '';
  severity = '';
  pageSize = 50;
  page = 0;
  sort: SortFieldEnum = null;
  sortDir: SortDirection = null;

  withDeviceName(name: string): this {
    this.deviceName = name;
    return this;
  }

  withAlertSummary(summary: string): this {
    this.alertSummary = summary;
    return this;
  }

  withStatus(status: string): this {
    this.status = status;
    return this;
  }

  withSeverity(severity: string): this {
    this.severity = severity;
    return this;
  }

  withPageSize(size: number): this {
    this.pageSize = size;
    return this;
  }

  withPage(page: number): this {
    this.page = page;
    return this;
  }

  withSort(sort: string, direction: SortDirection): this {
    this.sort = SortFieldEnum[sort];
    this.sortDir = direction;
    return this;
  }

  isMatchAllQuery(): boolean {
    return this.deviceName === '' && this.alertSummary === '' && this.status === '' && this.severity === '';
  }

  generateRequestObject(): object {
    return {
      'from': this.page * this.pageSize,
      'size': this.pageSize,
      'sort': this.generateSortObject(),
      'query': this.generateQueryObject()
    };
  }

  generateQueryObject(): object {
    if (this.isMatchAllQuery()) {
      return {
        'match_all': {}
      };
    }

    const boolQuery = {
      'bool': {
        'must': []
      }
    };
    if (this.deviceName !== '') {
      boolQuery.bool.must.push({
        'wildcard': {
          'ArrayDisplayName': `*${this.deviceName}*`
        }
      });
    }
    if (this.alertSummary !== '') {
      boolQuery.bool.must.push({
        'wildcard': {
          'Summary': `*${this.alertSummary}*`
        }
      });
    }

    if (this.status !== '') {
      boolQuery.bool.must.push({
        'term': {
          'State': this.status
        }
      });
    }

    if (this.severity !== '') {
      boolQuery.bool.must.push({
        'term': {
          'Severity': this.severity
        }
      });
    }

    return boolQuery;
  }

  generateSortObject(): object {
    if (!this.sort || !this.sortDir) {
      return {};
    }

    const toReturn = {};
    toReturn[this.sort] = {
      'order': `${this.sortDir}`
    };
    return toReturn;
  }
}
