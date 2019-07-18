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

package logger

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// CensorAPITokensFromFormatter takes the given formatter and adds an extra formatter
func CensorAPITokensFromFormatter(innerFormatter log.Formatter) log.Formatter {
	apiServerTokenRegex, _ := regexp.Compile("\\\\\"api_token\\\\\":\\\\\"(.*?)\\\\\"")
	apiServerTokenRegexCutOff, _ := regexp.Compile("\\\\\"api_token\\\\\":\\\\\"(.*?) \\.\\.\\.\\.\\.")
	return &tokenCensoringFormatter{nested: innerFormatter,
		apiServerTokenRegex:       apiServerTokenRegex,
		apiServerTokenRegexCutOff: apiServerTokenRegexCutOff,
	}
}

func (t *tokenCensoringFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := t.nested.Format(entry)
	// If it errored just return the original value
	if err != nil {
		return formatted, err
	}
	stringified := string(formatted)
	// Replace API tokens that appear in the format of \"api_token\":\"this-is-an-api-token\"
	stringified = t.apiServerTokenRegex.ReplaceAllString(stringified, "\\\"api_token\\\":\\\"****\\\"")
	// Replace API tokens that got cut off by ellipses: such as \"api_token\":\"this-is-an-api-tok ....."
	stringified = t.apiServerTokenRegexCutOff.ReplaceAllString(stringified, "\\\"api_token\\\":\\\"**** .....")
	return []byte(stringified), err
}
