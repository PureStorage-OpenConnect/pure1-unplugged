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

import { AfterViewInit, Component, Input, OnChanges, OnInit, QueryList, SimpleChanges,
          ViewChildren, ViewEncapsulation } from '@angular/core';
import * as d3 from 'd3';

@Component({
  selector: 'dial',
  template: '<svg #dialContainer></svg>',
  styleUrls: ['./dial.component.scss'],
  encapsulation: ViewEncapsulation.None,
})
export class DialComponent implements OnInit, AfterViewInit, OnChanges {

  @Input() size = 125;
  @Input() percentFull = 0.0;
  @Input() status: 'ok' | 'warning' | 'critical' = 'ok';
  @Input() centerLabel = '';
  @Input() labelSize = '12px';

  @ViewChildren('dialContainer') dialContainer: QueryList<any>;

  private dataArc: any;
  private backgroundArc: any;
  private isInitialized = false; // yes
  private svgElem: any; // yes

  constructor() { }

  ngOnInit() { }

  ngAfterViewInit() {
    this.initD3();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!this.isInitialized) { return; } // If this isn't initialized we can't do anything yet

    this.drawLabel();
    this.drawArcs();
  }

  private initD3(): void {
    this.svgElem = d3.select(this.dialContainer.first.nativeElement)
        .attr('width', this.size)
        .attr('height', this.size);

    this.drawLabel();

    // Draw background arc
    this.backgroundArc = d3.arc()
            .innerRadius(this.size / 2 - this.size * .12)
            .outerRadius(this.size / 2)
            .startAngle(0)
            .endAngle(Math.PI * 2);

    this.svgElem.append('path')
        .attr('d', this.backgroundArc)
        .attr('transform', 'translate(' + (this.size / 2) + ',' + (this.size / 2) + ')')
        .attr('class', 'background-arc');
    this.drawArcs();

    this.isInitialized = true;
  }

  private drawLabel(): void {
    this.svgElem.selectAll('.center-label').remove();

    this.svgElem.append('text')
        .attr('text-anchor', 'middle')
        .attr('y', '50%')
        .attr('x', '50%')
        .attr('dy', '0.35em')
        .attr('font-size', this.labelSize)
        .attr('fill', '#717171')
        .attr('class', 'center-label')
        .text(this.centerLabel);
  }

  private drawArcs(): void {
    this.svgElem.selectAll('.data-arc').remove();

    // Draw data arc
    this.dataArc = d3.arc()
                .innerRadius(this.size / 2 - this.size * .12)
                .outerRadius(this.size / 2)
                .startAngle(0)
                .endAngle(Math.PI * 2 * this.percentFull);

    this.svgElem.append('path')
            .attr('d', this.dataArc)
            .attr('transform', 'translate(' + (this.size / 2) + ',' + (this.size / 2) + ')')
            .attr('class', 'data-arc ' + this.status);
  }

}
