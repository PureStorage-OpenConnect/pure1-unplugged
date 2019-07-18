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
 * Provides some shared utilities to aid in the mocking process.
 */
export class DeviceMockUtil {
    static fakeIds: string[];
    static initialIdCount = 2;

    static generateFakeId(): string {
        const hexChars = '0123456789abcdef';
        const length = 24;
        let toReturn = '';
        for (let i = 0; i < length; i++) {
            toReturn += hexChars.charAt(Math.random() * hexChars.length);
        }
        return toReturn;
    }

    static initialize() {
        if (DeviceMockUtil.fakeIds) {
            return; // Already initialized
        }
        DeviceMockUtil.fakeIds = [];

        for (let i = 0; i < DeviceMockUtil.initialIdCount; i++) {
            DeviceMockUtil.fakeIds.push(DeviceMockUtil.generateFakeId());
        }
    }
}
