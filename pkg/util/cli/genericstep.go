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

package cli

import "github.com/sirupsen/logrus"

type genericStep struct {
	fn        func(ctx Context) error
	text      string
	stopOnErr bool
}

// Run the configured function, no error checking or logging
func (s *genericStep) Run(ctx Context) error {
	err := s.fn(ctx)
	if err != nil {
		if s.stopOnErr {
			return err
		}
		logrus.Errorf("workflow step failed, but is not terminating (will continue): %s", err.Error())
	}
	return nil
}

// Text just returns whatever was configured on the structs "text" field
func (s *genericStep) Text() string {
	return s.text
}

// Silent returns true iff there is text set for this step
func (s *genericStep) Silent() bool {
	return s.text == ""
}
