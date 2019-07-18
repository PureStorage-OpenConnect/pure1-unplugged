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

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MakeHTTPErr constructs an HTTPErr instance: this mainly
// makes it a bit easier to read and pass into return statements
func MakeHTTPErr(code int, err error) *HTTPErr {
	if httperr, ok := err.(*HTTPErr); ok {
		// This is already an HTTP error: let's just rewrap the inner one in the new code
		return &HTTPErr{Code: code, Inner: httperr.Inner}
	}

	return &HTTPErr{Code: code, Inner: err}
}

// MakeInternalHTTPErr is a wrapper for MakeHTTPErr that
// passes in http.StatusInternalServerError, since we use
// that response quite often
func MakeInternalHTTPErr(err error) *HTTPErr {
	return MakeHTTPErr(http.StatusInternalServerError, err)
}

// MakeBadRequestHTTPErr is a wrapper for MakeHTTPErr that
// passes in http.StatusBadRequest, since we use
// that response quite often
func MakeBadRequestHTTPErr(err error) *HTTPErr {
	return MakeHTTPErr(http.StatusBadRequest, err)
}

// HTTPErr.Error is simply a wrapper around
// the inner error's Error() method with a
// nil check
func (e *HTTPErr) Error() string {
	if e.Inner == nil {
		return ""
	}
	return e.Inner.Error()
}

// AssertIsHTTPErrOfCode performs 4 assertions:
// 0. err is actually an error
// 1. err is an HTTPErr instance
// 2. err.Code = the given code
// 3. err.Inner != nil
func AssertIsHTTPErrOfCode(t *testing.T, err error, code int) {
	assert.Error(t, err)
	assert.IsType(t, &HTTPErr{}, err)
	httpErr := err.(*HTTPErr)
	assert.Equal(t, code, httpErr.Code)
	assert.NotNil(t, httpErr.Inner)
}
