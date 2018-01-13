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

package templating

import (
	"testing"
)

func TestHasTemplating(t *testing.T) {
	type in struct {
		vars []string
	}
	type out struct {
		hasTemplating bool
	}
	tests := []struct {
		in  in
		out out
	}{
		{
			in{[]string{"foo=BAR"}},
			out{false},
		},
		{
			in{[]string{"foo={PUBLIC_IPV4}"}},
			out{true},
		},
		{
			in{[]string{"foo=PUBLIC_IPV4}"}},
			out{true},
		},
		{
			in{[]string{"foo={PUBLIC_IPV4"}},
			out{true},
		},
		{
			in{[]string{"foo=}PUBLIC_IPV4{"}},
			out{true},
		},
	}
	for i, test := range tests {
		if test.out.hasTemplating != HasTemplating(test.in.vars) {
			t.Errorf("#%d: hasTemplating didn't match", i)
		}
	}
}

func TestPerformTemplating(t *testing.T) {
	type in struct {
		platform string
		vars     []string
	}
	type out struct {
		vars []string
		err  error
	}
	tests := []struct {
		in  in
		out out
	}{
		{
			in{platform: "aws"},
			out{err: ErrUnknownPlatform},
		},
		{
			in{platform: "ec2", vars: []string{"foo", "bar"}},
			out{vars: []string{"foo", "bar"}},
		},
		{
			in{platform: "ec2", vars: []string{"foo: {HOSTNAME}", "bar"}},
			out{vars: []string{"foo: ${COREOS_EC2_HOSTNAME}", "bar"}},
		},
		{
			in{platform: "digitalocean", vars: []string{"foo: {PRIVATE_IPV4}", "bar"}},
			out{vars: []string{"foo: ${COREOS_DIGITALOCEAN_IPV4_PRIVATE_0}", "bar"}},
		},
		{
			in{platform: "digitalocean", vars: []string{"foo: {PRIVATE_IPV4} {PUBLIC_IPV4}", "bar"}},
			out{vars: []string{"foo: ${COREOS_DIGITALOCEAN_IPV4_PRIVATE_0} ${COREOS_DIGITALOCEAN_IPV4_PUBLIC_0}", "bar"}},
		},
		{
			in{platform: "azure", vars: []string{"foo: }HOSTNAME{", "bar"}},
			out{vars: []string{"foo: }HOSTNAME{", "bar"}},
		},
		{
			in{platform: "packet", vars: []string{"foo: {BAZ}", "bar"}},
			out{err: ErrUnknownField},
		},
	}
	for i, test := range tests {
		outVars, err := PerformTemplating(test.in.platform, test.in.vars)
		if err != test.out.err {
			t.Errorf("#%d: err (%v) didn't match expectedErr (%v)", i, err, test.out.err)
			continue
		}
		if err != nil {
			continue
		}
		if len(outVars) != len(test.in.vars) {
			t.Errorf("#%d: length of vars changed, was %d and is now %d", i, len(test.in.vars), len(outVars))
			continue
		}
		for j := range outVars {
			if test.out.vars[j] != outVars[j] {
				t.Errorf("#%d: var %d didn't match expected result, got %q, expected %q", i, j, outVars[j], test.out.vars[j])
			}
		}
	}
}
