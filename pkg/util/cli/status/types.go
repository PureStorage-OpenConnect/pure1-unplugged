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

package status

// ShortMsg is just a status string wrapper
type ShortMsg string

// Define some common statuses
const (
	OK    ShortMsg = "OK"
	NotOK ShortMsg = "ERROR"
)

// Info holds status information
type Info struct {
	Name    string
	Ok      ShortMsg
	Details string
}

// CLIContext is a helper to be used as the CmdContext for status CLI workflows
type CLIContext struct {
	Statuses []Info
}
