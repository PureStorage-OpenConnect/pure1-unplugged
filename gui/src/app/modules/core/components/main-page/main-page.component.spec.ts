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

import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { HttpClient, HttpHandler } from '@angular/common/http';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { NgbTooltipModule } from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import { CookieService } from 'ngx-cookie-service';
import { DeviceAlertService } from '../../../device-alert/device-alert.service';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { MainPageComponent } from './main-page.component';

describe('MainPageComponent', () => {
  let component: MainPageComponent;
  let fixture: ComponentFixture<MainPageComponent>;
  let nativeElem: any = null;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ MainPageComponent ],
      imports: [ NgbTooltipModule.forRoot(), RouterTestingModule ],
      providers: [ SidebarComponent, CookieService, DeviceAlertService, HttpClient, HttpHandler ],
      schemas: [ NO_ERRORS_SCHEMA ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(MainPageComponent);
    component = fixture.componentInstance;
    nativeElem = fixture.debugElement.nativeElement;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should start expanded', () => {
    expect(nativeElem.querySelector('#page-wrapper').classList.contains('retracted')).toBeFalsy();
  });

  it('should have a sidebar', () => {
    expect(nativeElem.querySelector('#page-wrapper sidebar')).toBeTruthy();
  });
});

describe('TokenEmailParser', () => {
  it('should fail on an null token', () => {
    expect(MainPageComponent.getTokenEmail(null)).toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should fail on an undefined token', () => {
    expect(MainPageComponent.getTokenEmail(undefined)).toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should fail on an empty token', () => {
    expect(MainPageComponent.getTokenEmail('')).toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should fail on a malformed token (not enough segments)', () => {
    expect(MainPageComponent.getTokenEmail('a.a')).toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should fail on an invalid claims section (invalid Base64)', () => {
    expect(MainPageComponent.getTokenEmail('a.aaaaa.a')).toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should fail on an invalid claims section (invalid JSON)', () => {
    // "{{[--bad;json::{?" in Base64
    expect(MainPageComponent.getTokenEmail('a.e3tbLS1iYWQ7anNvbjo6ez8=.a')).toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should fail on an invalid claims section (missing email)', () => {
    // "{"not-email": "pureuser@purestorage.com"}" in Base64
    expect(MainPageComponent.getTokenEmail('a.eyJub3QtZW1haWwiOiAicHVyZXVzZXJAcHVyZXN0b3JhZ2UuY29tIn0=.a'))
    .toEqual(MainPageComponent.invalidTokenCookieMessage);
  });

  it('should succeed on a valid token', () => {
    // "{"email": "pureuser@purestorage.com"}" in Base64
    expect(MainPageComponent.getTokenEmail('a.eyJlbWFpbCI6ICJwdXJldXNlckBwdXJlc3RvcmFnZS5jb20ifQ==.a')).toEqual('pureuser@purestorage.com');
  });
});

describe('TokenExpirationParser', () => {
  it('should fail on a null token', () => {
    expect(MainPageComponent.getTokenExpiryTime(null)).toBeNull();
  });

  it('should fail on an undefined token', () => {
    expect(MainPageComponent.getTokenExpiryTime(undefined)).toBeNull();
  });

  it('should fail on an empty token', () => {
    expect(MainPageComponent.getTokenExpiryTime('')).toBeNull();
  });

  it('should fail on a malformed token (not enough segments)', () => {
    expect(MainPageComponent.getTokenExpiryTime('a.a')).toBeNull();
  });

  it('should fail on an invalid claims section (invalid Base64)', () => {
    expect(MainPageComponent.getTokenExpiryTime('a.aaaaa.a')).toBeNull();
  });

  it('should fail on an invalid claims section (invalid JSON)', () => {
    // "{{[--bad;json::{?" in Base64
    expect(MainPageComponent.getTokenExpiryTime('a.e3tbLS1iYWQ7anNvbjo6ez8=.a')).toBeNull();
  });

  it('should fail on an invalid claims section (missing exp)', () => {
    // "{"not-exp": 1561590164}" in Base64
    expect(MainPageComponent.getTokenExpiryTime('a.eyJub3QtZXhwIjoxNTYxNTkwMTY0fQ.a')).toBeNull();
  });

  it('should fail on an invalid claims section (non-numeric exp)', () => {
    // "{"exp":"asdf"}" in Base64
    expect(MainPageComponent.getTokenExpiryTime('a.eyJleHAiOiJhc2RmIn0=.a')).toBeNull();
  });

  it('should succeed on a valid token', () => {
    const claims = btoa(JSON.stringify({
      exp: moment().unix() + 5
    }));
    const token = `a.${claims}.a`;
    expect(MainPageComponent.getTokenExpiryTime(token)).toBe(5);
  });
});
