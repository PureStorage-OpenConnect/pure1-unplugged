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

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"sort"
)

// PrintTable will print a formatted table of the status Info's stored on the CLIContext.
func PrintTable(ctx CLIContext) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"Check", "Status", "Details"})

	var statusKeys []string
	statusMap := map[string]Info{}
	for _, curStatus := range ctx.Statuses {
		statusKeys = append(statusKeys, curStatus.Name)
		statusMap[curStatus.Name] = curStatus
	}

	// Sort the status output
	sort.Strings(statusKeys)

	errCount := 0
	for _, key := range statusKeys {
		curStatus := statusMap[key]
		if curStatus.Ok != OK {
			errCount++
		}
		table.Append([]string{curStatus.Name, string(curStatus.Ok), curStatus.Details})
	}
	table.Render()
	passedAlerts := len(ctx.Statuses) - errCount
	fmt.Printf("\nSuccessful Checks %d/%d\n\n", passedAlerts, len(ctx.Statuses))

	if errCount > 0 {
		fmt.Println("System has ERROR's")
		fmt.Println("See log for error details!")
	} else {
		fmt.Printf("System is OK\n\n")
	}
}
