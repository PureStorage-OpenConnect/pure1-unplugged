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

import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';

@Component({
  selector: 'pagination-indicator',
  templateUrl: './pagination-indicator.component.html',
  styleUrls: ['./pagination-indicator.component.scss']
})
export class PaginationIndicatorComponent implements OnInit {
  @Input() itemCount = 0;
  @Input() itemsPerPage = 0;
  @Input() currentPage = 0;

  @Output() pageChange = new EventEmitter<number>(); // emits new page index

  constructor() {
    this.currentPage = 0;
  }

  ngOnInit() { }

  getPageCount(): number {
    return Math.ceil((this.itemCount / this.itemsPerPage));
  }

  getCurrentPageEndingIndex(): number {
    if (this.currentPage >= this.getPageCount() - 1) {
      return this.itemCount;
    } else {
      return (this.currentPage + 1) * this.itemsPerPage;
    }
  }

  pageClick(direction: number) {
    const oldPage = this.currentPage;
    // Add direction to current page, clamping to [0, pageCount)
    this.currentPage = Math.min(Math.max(0, this.currentPage + direction), this.getPageCount() - 1);

    // Only emit in case of value change
    if (this.currentPage !== oldPage) {
      this.pageChange.emit(this.currentPage);
    }
  }
}
