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

func TestValidateWindowStart(t *testing.T) {
	type in struct {
		start string
	}
	type out struct {
		r report.Report
	}
	tests := []struct {
		in  in
		out out
	}{
		{in{"mon 00:00"}, out{report.Report{}}},
		{in{"tue 23:59"}, out{report.Report{}}},
		{in{"sun 12:00"}, out{report.Report{}}},
		{
			in{"mon00:00"},
			out{report.ReportFromError(ErrParsingWindow, report.EntryError)},
		},
		{
			in{"foo 00:00"},
			out{report.ReportFromError(ErrUnknownDay, report.EntryError)},
		},
		{
			in{"mon 0000"},
			out{report.ReportFromError(ErrParsingWindow, report.EntryError)},
		},
		{
			in{"mon 24:00"},
			out{report.ReportFromError(ErrParsingWindow, report.EntryError)},
		},
	}

	for i, test := range tests {
		r := Locksmith{WindowStart: test.in.start}.ValidateWindowStart()
		if !reflect.DeepEqual(test.out.r, r) {
			t.Errorf("#%d: wanted %v, got %v", i, test.out.r, r)
		}
	}
}
