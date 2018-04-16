// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/validate/report"
)

func TestValidateUpdateGroup(t *testing.T) {
	tests := []struct {
		in  string
		out report.Report
	}{
		{"stable", report.Report{}},
		{"beta", report.Report{}},
		{"alpha", report.Report{}},
		{
			"super-alpha",
			report.ReportFromError(ErrUnknownGroup, report.EntryWarning),
		},
	}

	for i, test := range tests {
		r := Update{Group: UpdateGroup(test.in)}.Validate()
		if !reflect.DeepEqual(test.out, r) {
			t.Errorf("#%d: wanted %v, got %v", i, test.out, r)
		}
	}
}
