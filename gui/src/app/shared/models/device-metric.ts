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

export class DeviceMetric {
    // Metric metadata:
    ArrayID: string;
    ArrayName: string;
    CreatedAt: number;
    DisplayName: string;

    VolumeCount: number;
    FileSystemCount: number;
    SnapshotCount: number;
    VolumePendingEradicationCount: number;

    UsedSpace: number;
    TotalSpace: number;
    PercentFull: number;

    SharedSpace: number;
    SnapshotSpace: number;
    SystemSpace: number;
    VolumeSpace: number;

    BytesPerOp: number;
    BytesPerRead: number;
    BytesPerWrite: number;

    WriteBandwidth: number;
    ReadBandwidth: number;

    ReadIOPS: number;
    WriteIOPS: number;
    OtherIOPS: number;

    DataReduction: number;
    TotalReduction: number;

    ReadLatency: number;
    WriteLatency: number;
    OtherLatency: number;

    HostCount: number;
    QueueDepth: number;
    AlertMessageCount: number;

    Tags: object;
}
