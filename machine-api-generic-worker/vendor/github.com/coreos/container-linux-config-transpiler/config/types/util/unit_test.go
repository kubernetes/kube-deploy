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

package util

import (
	"testing"
)

type intype struct {
	unit    []string
	service []string
	install []string
}

func TestBuildUnit(t *testing.T) {
	tests := []struct {
		in  intype
		out string
	}{
		{
			in: intype{
				unit:    []string{},
				service: []string{},
				install: []string{},
			},
			out: "",
		},
		{
			in: intype{
				unit: []string{
					"Description=etcd (System Application Container)",
				},
				service: []string{},
				install: []string{},
			},
			out: `[Unit]
Description=etcd (System Application Container)`,
		},
		{
			in: intype{
				unit: []string{
					"Description=etcd (System Application Container)",
					"Documentation=https://github.com/coreos/etcd",
				},
				service: []string{},
				install: []string{},
			},
			out: `[Unit]
Description=etcd (System Application Container)
Documentation=https://github.com/coreos/etcd`,
		},
		{
			in: intype{
				unit: []string{
					"Description=etcd (System Application Container)",
					"Documentation=https://github.com/coreos/etcd",
				},
				service: []string{
					"Environment=\"ETCD_IMAGE_TAG=v3.0.10\"",
				},
				install: []string{},
			},
			out: `[Unit]
Description=etcd (System Application Container)
Documentation=https://github.com/coreos/etcd

[Service]
Environment="ETCD_IMAGE_TAG=v3.0.10"`,
		},
		{
			in: intype{
				unit: []string{
					"Description=etcd (System Application Container)",
					"Documentation=https://github.com/coreos/etcd",
				},
				service: []string{},
				install: []string{
					"WantedBy=multi-user.target",
				},
			},
			out: `[Unit]
Description=etcd (System Application Container)
Documentation=https://github.com/coreos/etcd

[Install]
WantedBy=multi-user.target`,
		},
		{
			in: intype{
				unit: []string{
					"Description=etcd (System Application Container)",
					"Documentation=https://github.com/coreos/etcd",
				},
				service: []string{
					"Environment=\"ETCD_IMAGE_TAG=v3.0.10\"",
				},
				install: []string{
					"WantedBy=multi-user.target",
				},
			},
			out: `[Unit]
Description=etcd (System Application Container)
Documentation=https://github.com/coreos/etcd

[Service]
Environment="ETCD_IMAGE_TAG=v3.0.10"

[Install]
WantedBy=multi-user.target`,
		},
	}

	for i, test := range tests {
		unit := NewSystemdUnit()
		for _, l := range test.in.unit {
			unit.Unit.Add(l)
		}
		for _, l := range test.in.service {
			unit.Service.Add(l)
		}
		for _, l := range test.in.install {
			unit.Install.Add(l)
		}
		res := unit.String()
		if res != test.out {
			t.Errorf("#%d: result didn't match expected output.\nResult:\n%s\n\nExpected:\n%s", i, res, test.out)
		}
	}
}
