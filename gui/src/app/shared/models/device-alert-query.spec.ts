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

import { DeviceAlertQuery } from './device-alert-query';

describe('DeviceAlertQuery', () => {
  it('should match all on empty query', () => {
    const query = new DeviceAlertQuery();

    expect(query.generateQueryObject()).toEqual({
      'match_all': {}
    });
  });

  it('should match wildcard name with just name', () => {
    const query = new DeviceAlertQuery().withDeviceName('asdf');

    expect(query.generateQueryObject()).toEqual({
      'bool': {
        'must': [
          {
            'wildcard': { 'ArrayDisplayName': '*asdf*' }
          }
        ]
      }
    });
  });

  it('should match wildcard summary with just alert summary', () => {
    const query = new DeviceAlertQuery().withAlertSummary('some problem');

    expect(query.generateQueryObject()).toEqual({
      'bool': {
        'must': [
          {
            'wildcard': { 'Summary': '*some problem*' }
          }
        ]
      }
    });
  });

  it('should match exact status with just status', () => {
    const query = new DeviceAlertQuery().withStatus('closed');

    expect(query.generateQueryObject()).toEqual({
      'bool': {
        'must': [
          {
            'term': { 'State': 'closed' }
          }
        ]
      }
    });
  });

  it('should match exact severity with just severity', () => {
    const query = new DeviceAlertQuery().withSeverity('critical');

    expect(query.generateQueryObject()).toEqual({
      'bool': {
        'must': [
          {
            'term': { 'Severity': 'critical' }
          }
        ]
      }
    });
  });

  it('should match all 4 queries if all 4 are set', () => {
    const query = new DeviceAlertQuery()
                    .withDeviceName('asdf')
                    .withAlertSummary('some problem')
                    .withStatus('closed')
                    .withSeverity('critical');

    expect(query.generateQueryObject()).toEqual({
      'bool': {
        'must': [
          {
            'wildcard': { 'ArrayDisplayName': '*asdf*'}
          },
          {
            'wildcard': { 'Summary': '*some problem*'}
          },
          {
            'term': { 'State': 'closed'}
          },
          {
            'term': { 'Severity': 'critical'}
          }
        ]
      }
    });
  });

  it('should return an empty string with no sort parameters set', () => {
    const query = new DeviceAlertQuery().withSort('', null);

    expect(query.generateSortObject()).toEqual({});
  });

  it('should return an empty string with the sort direction set to null', () => {
    const query = new DeviceAlertQuery().withSort('device_name', null);

    expect(query.generateSortObject()).toEqual({});
  });

  it('should return the correct field for device_name', () => {
    const query = new DeviceAlertQuery().withSort('device_name', 'asc');

    expect(query.generateSortObject()).toEqual({
      'ArrayDisplayName': {
        'order': 'asc'
      }
    });
  });

  it('should return the correct field for severity', () => {
    const query = new DeviceAlertQuery().withSort('severity', 'asc');

    expect(query.generateSortObject()).toEqual({
      'SeverityIndex': {
        'order': 'asc'
      }
    });
  });

  it('should return the correct field for status', () => {
    const query = new DeviceAlertQuery().withSort('state', 'asc');

    expect(query.generateSortObject()).toEqual({
      'State': {
        'order': 'asc'
      }
    });
  });

  it('should return the correct field for summary', () => {
    const query = new DeviceAlertQuery().withSort('summary', 'asc');

    expect(query.generateSortObject()).toEqual({
      'Summary.raw': {
        'order': 'asc'
      }
    });
  });

  it('should return the correct field for updated', () => {
    const query = new DeviceAlertQuery().withSort('created', 'asc');

    expect(query.generateSortObject()).toEqual({
      'Created': {
        'order': 'asc'
      }
    });
  });
});
