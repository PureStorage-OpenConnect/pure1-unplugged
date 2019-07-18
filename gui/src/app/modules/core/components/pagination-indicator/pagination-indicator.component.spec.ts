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

import { HttpClientModule } from '@angular/common/http';
import { InlineSVGModule } from 'ng-inline-svg';
import { PaginationIndicatorComponent } from './pagination-indicator.component';

function getCurrentCountText(fixture: ComponentFixture<PaginationIndicatorComponent>): string {
  const currentCountElem = fixture.elementRef.nativeElement.querySelectorAll('.current-count');
  if (currentCountElem.length === 0) {
    return undefined;
  }
  return currentCountElem[0].innerHTML;
}

function getTotalCountText(fixture: ComponentFixture<PaginationIndicatorComponent>): string {
  const totalCountElem = fixture.elementRef.nativeElement.querySelectorAll('.total-count');
  if (totalCountElem.length === 0) {
    return undefined;
  }
  return totalCountElem[0].innerHTML;
}

function getButtonEnabled(fixture: ComponentFixture<PaginationIndicatorComponent>, index: number): boolean {
  const buttons = fixture.elementRef.nativeElement.querySelectorAll('button');
  if (buttons.length !== 2) {
    return undefined;
  }
  return !buttons[index].disabled;
}

describe('PaginationIndicatorComponent', () => {
  let component: PaginationIndicatorComponent;
  let fixture: ComponentFixture<PaginationIndicatorComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PaginationIndicatorComponent ],
      imports: [
        InlineSVGModule.forRoot(),
        HttpClientModule
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PaginationIndicatorComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should have both buttons disabled if length is set to 0', () => {
    component.itemCount = 0;
    // Set some other random values for testing
    component.currentPage = 0;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const backwardEnabled = getButtonEnabled(fixture, 0);
    expect(backwardEnabled).toBeDefined();
    expect(backwardEnabled).toBeFalsy();
    const forwardEnabled = getButtonEnabled(fixture, 1);
    expect(forwardEnabled).toBeDefined();
    expect(forwardEnabled).toBeFalsy();
  });

  it('should have both buttons disabled if length is set to the page size', () => {
    component.itemCount = 25;
    // Set some other random values for testing
    component.currentPage = 0;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const backwardEnabled = getButtonEnabled(fixture, 0);
    expect(backwardEnabled).toBeDefined();
    expect(backwardEnabled).toBeFalsy();
    const forwardEnabled = getButtonEnabled(fixture, 1);
    expect(forwardEnabled).toBeDefined();
    expect(forwardEnabled).toBeFalsy();
  });

  it('should have the forward button enabled if it can go forwards', () => {
    component.itemCount = 26; // Just over one page
    // Set some other random values for testing
    component.currentPage = 0;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const backwardEnabled = getButtonEnabled(fixture, 0);
    expect(backwardEnabled).toBeDefined();
    expect(backwardEnabled).toBeFalsy();
    const forwardEnabled = getButtonEnabled(fixture, 1);
    expect(forwardEnabled).toBeDefined();
    expect(forwardEnabled).toBeTruthy();
  });

  it('should have the backward button enabled if it can go backwards', () => {
    component.itemCount = 26; // Just over one page
    component.currentPage = 1; // So we can go back
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const backwardEnabled = getButtonEnabled(fixture, 0);
    expect(backwardEnabled).toBeDefined();
    expect(backwardEnabled).toBeTruthy();
    const forwardEnabled = getButtonEnabled(fixture, 1);
    expect(forwardEnabled).toBeDefined();
    expect(forwardEnabled).toBeFalsy();
  });

  it('should have the both buttons enabled if it can go both directions', () => {
    component.itemCount = 51; // > 2 pages
    component.currentPage = 1;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const backwardEnabled = getButtonEnabled(fixture, 0);
    expect(backwardEnabled).toBeDefined();
    expect(backwardEnabled).toBeTruthy();
    const forwardEnabled = getButtonEnabled(fixture, 1);
    expect(forwardEnabled).toBeDefined();
    expect(forwardEnabled).toBeTruthy();
  });

  it('should show 0-0 of 0 if item count is set to 0', () => {
    component.itemCount = 0;
    // Set some other random values for testing
    component.currentPage = 3;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const currentCountText = getCurrentCountText(fixture);
    expect(currentCountText).toBeDefined();
    expect(currentCountText).toBe('0-0');
    const totalCountText = getTotalCountText(fixture);
    expect(totalCountText).toBeDefined();
    expect(totalCountText).toBe(' of 0');
  });

  it('should display the proper page values if it\'s at the beginning with an incomplete page', () => {
    component.itemCount = 13; // Just over one page
    component.currentPage = 0;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const currentCountText = getCurrentCountText(fixture);
    expect(currentCountText).toBeDefined();
    expect(currentCountText).toBe('1-13');
    const totalCountText = getTotalCountText(fixture);
    expect(totalCountText).toBeDefined();
    expect(totalCountText).toBe(' of 13');
  });

  it('should display the proper page values if it\'s at the beginning', () => {
    component.itemCount = 26; // Just over one page
    component.currentPage = 0;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const currentCountText = getCurrentCountText(fixture);
    expect(currentCountText).toBeDefined();
    expect(currentCountText).toBe('1-25');
    const totalCountText = getTotalCountText(fixture);
    expect(totalCountText).toBeDefined();
    expect(totalCountText).toBe(' of 26');
  });

  it('should display the proper page values if it\'s in the middle', () => {
    component.itemCount = 51; // 3 pages worth
    component.currentPage = 1; // 2nd page
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const currentCountText = getCurrentCountText(fixture);
    expect(currentCountText).toBeDefined();
    expect(currentCountText).toBe('26-50');
    const totalCountText = getTotalCountText(fixture);
    expect(totalCountText).toBeDefined();
    expect(totalCountText).toBe(' of 51');
  });

  it('should display the proper page values if it\'s at the end', () => {
    component.itemCount = 27; // More than just over one page
    component.currentPage = 1;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const currentCountText = getCurrentCountText(fixture);
    expect(currentCountText).toBeDefined();
    expect(currentCountText).toBe('26-27');
    const totalCountText = getTotalCountText(fixture);
    expect(totalCountText).toBeDefined();
    expect(totalCountText).toBe(' of 27');
  });

  it('should display the proper page values if it\'s at the end with a 1 item page', () => {
    component.itemCount = 26;
    component.currentPage = 1;
    component.itemsPerPage = 25;
    fixture.detectChanges();
    const currentCountText = getCurrentCountText(fixture);
    expect(currentCountText).toBeDefined();
    expect(currentCountText).toBe('26-26');
    const totalCountText = getTotalCountText(fixture);
    expect(totalCountText).toBeDefined();
    expect(totalCountText).toBe(' of 26');
  });
});
