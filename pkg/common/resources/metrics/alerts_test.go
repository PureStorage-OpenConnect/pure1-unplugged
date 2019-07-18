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

package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPopulateSeverityIndexEmpty(t *testing.T) {
	alert := Alert{
		Severity: "",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(0), alert.SeverityIndex)
}

func TestPopulateSeverityIndexInvalid(t *testing.T) {
	alert := Alert{
		Severity: "notaseverity",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(0), alert.SeverityIndex)
}

func TestPopulateSeverityIndexInfo(t *testing.T) {
	alert := Alert{
		Severity: "info",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(1), alert.SeverityIndex)
}

func TestPopulateSeverityIndexWarning(t *testing.T) {
	alert := Alert{
		Severity: "warning",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(2), alert.SeverityIndex)
}

func TestPopulateSeverityIndexCritical(t *testing.T) {
	alert := Alert{
		Severity: "critical",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(3), alert.SeverityIndex)
}

func TestPopulateSeverityIndexInfoUppercase(t *testing.T) {
	alert := Alert{
		Severity: "Info",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(1), alert.SeverityIndex)
}

func TestPopulateSeverityIndexWarningUppercase(t *testing.T) {
	alert := Alert{
		Severity: "Warning",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(2), alert.SeverityIndex)
}

func TestPopulateSeverityIndexCriticalUppercase(t *testing.T) {
	alert := Alert{
		Severity: "Critical",
	}
	alert.PopulateSeverityIndex()

	assert.Equal(t, byte(3), alert.SeverityIndex)
}
