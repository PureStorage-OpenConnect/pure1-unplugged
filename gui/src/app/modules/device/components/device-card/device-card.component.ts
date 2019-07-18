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
 * This component is a "card" for a storage device. It should display the name and other
 * relevant information.
 */
import { Component, Input, OnDestroy, OnInit } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import { ISlimScrollOptions } from 'ngx-slimscroll';
import { of, Subject } from 'rxjs';
import { catchError, takeUntil } from 'rxjs/operators';
import { DeviceAlert } from '../../../../shared/models/device-alert';
import { DeviceMetric } from '../../../../shared/models/device-metric';
import { DeviceTag } from '../../../../shared/models/device-tag';
import { StorageDevice } from '../../../../shared/models/storage-device';
import { DeviceMetricsService } from '../../../../shared/services/device-metrics.service';
import { DeviceTagsService } from '../../../../shared/services/device-tags.service';
import { DeviceService } from '../../device.service';
import { DeviceDeleteModalComponent } from '../device-delete-modal/device-delete-modal.component';
import { TagEditModalComponent } from '../tag-edit-modal/tag-edit-modal.component';

@Component({
  selector: 'device-card',
  templateUrl: './device-card.component.html',
  styleUrls: ['./device-card.component.scss']
})
export class DeviceCardComponent implements OnInit, OnDestroy {
  flipped = false;
  metricsStopService: Subject<void>;

  @Input() device: StorageDevice;
  @Input() tags: Map<string, Map<string, string>>;
  @Input() alerts: DeviceAlert[];

  latestMetric: DeviceMetric = null;
  dataPercentFull = 0.0;
  dataPercentFullRounded = 0;
  dataUsageString = '';
  dataStatus: 'ok' | 'warning' | 'critical' = 'ok';
  iopsString = '';
  iopsSuffix = '';
  bandwidthString = '';
  bandwidthSuffix = '';

  scrollOptions: ISlimScrollOptions = {
    alwaysVisible: true,
    position: 'right',

    gridOpacity: '0.2',
    gridBackground: '#333333',
    gridWidth: '7',
    gridBorderRadius: '7',
    gridMargin: '0',

    barOpacity: '0.4',
    barBackground: '#e6e6e6',
    barWidth: '7',
    barBorderRadius: '7',
    barMargin: '0',
  };

  static getRoundedPercentage(decimalFull: number): number {
    return Math.floor(decimalFull * 100);
  }

  constructor(private deviceService: DeviceService, private modalService: NgbModal, private tagsService: DeviceTagsService,
    private metricsService: DeviceMetricsService) { }

  ngOnInit() {
    console.debug(`Initializing device display for device "${this.device.name}"`);
    this.metricsStopService = new Subject<void>();
    this.metricsService.getLatestMetric(this.device.id).pipe(takeUntil(this.metricsStopService)).subscribe(metric => {
      this.processMetric(metric);
    });
  }

  processMetric(metric: DeviceMetric) {
    this.latestMetric = metric;
    this.dataPercentFull = metric.PercentFull;
    this.dataPercentFullRounded = DeviceCardComponent.getRoundedPercentage(this.dataPercentFull);
    if (this.dataPercentFull >= 1.0) {
      this.dataStatus = 'critical';
    } else if (this.dataPercentFull >= 0.9) {
      this.dataStatus = 'warning';
    } else {
      this.dataStatus = 'ok';
    }
    this.createTotalSpaceString();
    this.createIOPSStringAndSuffix();
    this.createBandwidthStringAndSuffix();
  }

  ngOnDestroy(): void {
    if (this.metricsStopService) {
      this.metricsStopService.next();
    }
  }

  createTotalSpaceString() {
    const totalSpacePower = this.getDiskSuffixPower(this.latestMetric.TotalSpace);
    const usedReduced = this.latestMetric.UsedSpace / Math.pow(1024, totalSpacePower);
    const totalReduced = this.latestMetric.TotalSpace / Math.pow(1024, totalSpacePower);
    this.dataUsageString = `${usedReduced.toFixed(1)} / ${totalReduced.toFixed(1)} ${this.getSuffix(totalSpacePower)}`;
  }

  createIOPSStringAndSuffix() {
    const iops = this.latestMetric.ReadIOPS + this.latestMetric.WriteIOPS;
    const iopsSuffixPower = this.getNumberSuffixPower(iops);
    const convertedIOPS = iops / Math.pow(1000, iopsSuffixPower);

    if (iopsSuffixPower > 0) {
      this.iopsString = convertedIOPS.toFixed(1);
    } else {
      this.iopsString = convertedIOPS.toFixed(0);
    }
    this.iopsSuffix = this.getSuffix(iopsSuffixPower);
  }

  createBandwidthStringAndSuffix() {
    const bandwidth = this.latestMetric.WriteBandwidth + this.latestMetric.ReadBandwidth;
    const bandwidthSuffixPower = this.getDiskSuffixPower(bandwidth);
    const convertedBandwidth = bandwidth / Math.pow(1024, bandwidthSuffixPower);

    // Bandwidth always has one decimal
    this.bandwidthString = convertedBandwidth.toFixed(1);
    this.bandwidthSuffix = this.getSuffix(bandwidthSuffixPower) + 'B/s';
  }

  getNumberSuffixPower(iops: number): number {
    if (iops === 0) { return 0; }

    // Use change of base formula to calculate log base 1000
    // Base 1000 here since this is IOPS, a number, not a disk space
    return Math.floor(Math.log(iops) / Math.log(1000));
  }

  getDiskSuffixPower(size: number): number {
    if (size === 0) { return 0; } // This is almost certainly never gonna happen, but just in case

    // Use change of base formula to calculate log base 1024
    // Base 1024 here because it's a disk space unit, not a generic number
    return Math.floor(Math.log(size) / Math.log(1024));
  }

  getSuffix(power: number): string {
    switch (power) {
      case 5:
        return 'P';
      case 4:
        return 'T';
      case 3:
        return 'G';
      case 2:
        return 'M';
      case 1:
        return 'K';
      case 0:
        return '';
    }
  }

  // Should never return null
  safeGetTag(ns: string, key: string): string {
    if (this.tags && this.tags.has(ns) && this.tags.get(ns).has(key)) {
      return this.tags.get(ns).get(key);
    }
    return '';
  }

  isFlashBlade(): boolean {
    return this.device.device_type === 'FlashBlade';
  }

  getTags(): DeviceTag[] {
    if (!this.tags) { return []; }

    const toReturn = [];

    this.tags.forEach((keyValue: Map<string, string>, namespace: string) => {
      keyValue.forEach((value: string, key: string) => {
        toReturn.push({ namespace: namespace, key: key, value: value });
      });
    });

    return toReturn;
  }

  getAsOfText(): string {
    const parsed = moment(this.device._as_of, ['YYYY-MM-DD[T]HH:mm:ss[Z]', 'YYYY-MM-DD[T]HH:mm:ss.SSS[Z]']);
    if (parsed.year() === 1) {
      // This is a zero date
      return 'never';
    }

    const now = moment();

    // Longer than a month, just return the date
    // Reasoning: this way the user will focus more on the date and that
    // "this was a long time ago", instead of getting bogged down looking
    // at an actual time (and is additionally cleaner). Contrast this with
    // something a bit more recent, where they may still be looking at logs
    // and determining what happened and a time would be useful to see more easily.
    if (now.diff(parsed, 'months', false) >= 1) {
      return parsed.format('DD MMM YYYY');
    }

    return parsed.format('DD MMM YYYY HH:mm:ss');
  }

  onFlipClick(): void {
    this.flipped = !this.flipped;
  }

  tagIconClick(): void {
    const activeModal = this.modalService.open(TagEditModalComponent, { backdrop: 'static', keyboard: false });
    activeModal.componentInstance.device = this.device;

    const tagsList = [];

    this.tags.forEach((nsMap, ns) => {
      nsMap.forEach((value, key) => {
        tagsList.push({ key: key, value: value, namespace: ns });
      });
    });

    activeModal.componentInstance.tags = tagsList;

    activeModal.result.then((output: DeviceTag[]) => {
      const userTags: { key: string, value: string, namespace: string }[] = <{ key: string, value: string, namespace: string }[]>[];

      output.filter(tag => tag.namespace === 'pure1-unplugged').forEach(tag => {
        userTags.push({ key: tag.key, value: tag.value, namespace: 'pure1-unplugged' });
      });

      // Create the patch to send
      this.tagsService.patch(this.device.id, userTags).subscribe(patchResult => {
        // Find any tags that we removed that no longer exist and remove them.
        // This is inside the subscribe block to prevent a current API server race condition
        // bug: see NSTK-911
        if (!this.tags.has('pure1-unplugged')) { return; }

        const keysToDelete = [];

        this.tags.get('pure1-unplugged').forEach((_, key) => {
          if (!userTags.some(t => t.key.localeCompare(key) === 0)) { // Key exists in current tags but not in patched tags: delete it!
            keysToDelete.push(key);
          }
        });
        this.tagsService.deleteTag(this.device.id, keysToDelete.join(',')).pipe(catchError(err => {
          console.error(err);
          return of(null);
        })).subscribe(_ => { });
      });
    }).catch((reason: any) => {

    });
  }

  deleteIconClick() {
    const openModal = this.modalService.open(DeviceDeleteModalComponent, { backdrop: 'static', keyboard: false });
    openModal.componentInstance.deviceName = this.device.name;
    openModal.result.then(() => {
      // DELETE THE DEVICE
      this.deviceService.delete(this.device.id).subscribe(x => { });
    }).catch(() => {
      // Do nothing, they cancelled
    });
  }

  getHighestSeverity(): string {
    let highestSeverity = 'none'; // 'none', 'info', 'warning', 'critical'

    this.alerts.forEach(alert => {
      switch (alert.Severity) {
        case 'info': {
          if (highestSeverity === 'none') {
            highestSeverity = 'info';
          }
          break;
        }
        case 'warning': {
          if (highestSeverity === 'none' || highestSeverity === 'info') {
            highestSeverity = 'warning';
          }
          break;
        }
        case 'critical': {
          if (highestSeverity !== 'critical') {
            highestSeverity = 'critical';
          }
          break;
        }
      }
    });

    return highestSeverity;
  }

  trackByNSKey(index, item) {
    return item.namespace + ':' + item.key; // Make a consistent value to track by
  }

  trackAlertByID(index, item: DeviceAlert) {
    return item.ArrayID + '-' + item.AlertID;
  }
}
