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

package errors

// JSONErr provides a common error struct for JSON
type JSONErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

// HTTPErr provides a basic struct to pass up which HTTP status code
// should be used with this error
type HTTPErr struct {
	Code  int   // The HTTP error code to use for this error
	Inner error // The actual error of this HTTPErr struct
}
