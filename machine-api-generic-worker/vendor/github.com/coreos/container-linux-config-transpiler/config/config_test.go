// Copyright 2016 CoreOS, Inc.
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

package config

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coreos/container-linux-config-transpiler/config/types"
	"github.com/coreos/container-linux-config-transpiler/internal/util"
	"github.com/coreos/go-semver/semver"
	ignTypes "github.com/coreos/ignition/config/v2_1/types"
	"github.com/coreos/ignition/config/validate/report"
)

func TestParse(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		cfg types.Config
		r   report.Report
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in: in{data: ``},
			out: out{
				cfg: types.Config{},
				r: report.Report{
					Entries: []report.Entry{{
						Kind:    report.EntryWarning,
						Message: "Configuration is empty",
					}},
				},
			},
		},
		{
			in: in{data: `
networkd:
  units:
    - name: bad.blah
      contents: not valid
`},
			out: out{cfg: types.Config{
				Networkd: types.Networkd{
					Units: []types.NetworkdUnit{
						{Name: "bad.blah", Contents: "not valid"},
					},
				},
			}},
		},

		// Timeouts
		{
			in: in{data: `
ignition:
  timeouts:
    http_response_headers: 30
    http_total: 31
`},
			out: out{cfg: types.Config{
				Ignition: types.Ignition{
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(30),
						HTTPTotal:           util.IntToPtr(31),
					},
				},
			}},
		},

		// Config
		{
			in: in{data: `
ignition:
  config:
    append:
      - source: http://example.com/test1
        verification:
          hash:
            function: sha512
            sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
      - source: http://example.com/test2
    replace:
      source: http://example.com/test3
      verification:
        hash:
          function: sha512
          sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
`},
			out: out{cfg: types.Config{
				Ignition: types.Ignition{
					Config: types.IgnitionConfig{
						Append: []types.ConfigReference{
							{
								Source: "http://example.com/test1",
								Verification: types.Verification{
									Hash: types.Hash{
										Function: "sha512",
										Sum:      "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
									},
								},
							},
							{
								Source: "http://example.com/test2",
							},
						},
						Replace: &types.ConfigReference{
							Source: "http://example.com/test3",
							Verification: types.Verification{
								Hash: types.Hash{
									Function: "sha512",
									Sum:      "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
								},
							},
						},
					},
				},
			}},
		},

		// Storage
		{
			in: in{data: `
storage:
  disks:
    - device: /dev/sda
      wipe_table: true
      partitions:
        - label: ROOT
          number: 7
          size: 100MB
          start: 50MB
          guid: 22222222-2222-2222-2222-222222222222
          type_guid: 11111111-1111-1111-1111-111111111111
        - label: DATA
          number: 12
          size: 1GB
          start: 300MB
          guid: 33333333-3333-3333-3333-333333333333
          type_guid: 00000000-0000-0000-0000-000000000000
        - label: NOTHING
        - label: ROOT_ON_RAID
          type_guid: raid_containing_root
          number: 13
        - label: SWAP
          number: 14
          type_guid: swap_partition
        - label: RAID
          number: 15
          type_guid: raid_partition
        - label: LINUX_FS
          number: 16
          type_guid: linux_filesystem_data
    - device: /dev/sdb
      wipe_table: true
  raid:
    - name: fast
      level: raid0
      devices:
        - /dev/sdc
        - /dev/sdd
    - name: durable
      level: raid1
      devices:
        - /dev/sde
        - /dev/sdf
        - /dev/sdg
      spares: 1
  filesystems:
    - name: filesystem1
      mount:
        device: /dev/disk/by-partlabel/ROOT
        format: btrfs
        create:
          force: true
          options:
            - -L
            - ROOT
    - name: filesystem2
      mount:
        device: /dev/disk/by-partlabel/DATA
        format: ext4
    - name: filesystem3
      path: /sysroot
    - name: filesystem4
      mount:
        device: /dev/disk/by-partlabel/DATA2
        format: xfs
        wipe_filesystem: true
        label: data2
        uuid: a51034e6-26b3-48df-beed-220562ac7ad1
        options:
          - "i'm an option"
          - "me too"
    - name: filesystem5
      mount:
        device: /dev/disk/by-partlabel/DATA3
        format: vfat
    - name: filesystem6
      mount:
        device: /dev/disk/by-partlabel/DATA4
        format: swap
  files:
    - path: /opt/file1
      filesystem: filesystem1
      contents:
        inline: file1
      mode: 0644
      user:
        id: 500
      group:
        id: 501
    - path: /opt/file2
      filesystem: filesystem1
      contents:
        remote:
          url: http://example.com/file2
          compression: gzip
          verification:
            hash:
              function: sha512
              sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
      mode: 0644
      user:
        id: 502
      group:
        id: 503
    - path: /opt/file3
      filesystem: filesystem2
      contents:
        remote:
          url: http://example.com/file3
          compression: gzip
      mode: 0400
      user:
        id: 1000
      group:
        id: 1001
    - path: /opt/file4
      filesystem: filesystem2
  directories:
    - path: /opt/dir1
      filesystem: filesystem1
      mode: 0755
      user:
        name: core
      group:
        name: core
  links:
    - path: /opt/link1
      filesystem: filesystem1
      user:
        name: noone
      group:
        name: systemd-journald
      hard: false
      target: /opt/file2
    - path: /opt/link2
      filesystem: filesystem2
      hard: true
      target: /opt/file3
`},
			out: out{cfg: types.Config{
				Storage: types.Storage{
					Disks: []types.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []types.Partition{
								{
									Label:    "ROOT",
									Number:   7,
									Size:     "100MB",
									Start:    "50MB",
									GUID:     "22222222-2222-2222-2222-222222222222",
									TypeGUID: "11111111-1111-1111-1111-111111111111",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     "1GB",
									Start:    "300MB",
									GUID:     "33333333-3333-3333-3333-333333333333",
									TypeGUID: "00000000-0000-0000-0000-000000000000",
								},
								{
									Label: "NOTHING",
								},
								{
									Label:    "ROOT_ON_RAID",
									Number:   13,
									TypeGUID: "raid_containing_root",
								},
								{
									Label:    "SWAP",
									Number:   14,
									TypeGUID: "swap_partition",
								},
								{
									Label:    "RAID",
									Number:   15,
									TypeGUID: "raid_partition",
								},
								{
									Label:    "LINUX_FS",
									Number:   16,
									TypeGUID: "linux_filesystem_data",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
					Arrays: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []string{"/dev/sdc", "/dev/sdd"},
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []string{"/dev/sde", "/dev/sdf", "/dev/sdg"},
							Spares:  1,
						},
					},
					Filesystems: []types.Filesystem{
						{
							Name: "filesystem1",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.Create{
									Force:   true,
									Options: []string{"-L", "ROOT"},
								},
							},
						},
						{
							Name: "filesystem2",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA",
								Format: "ext4",
							},
						},
						{
							Name: "filesystem3",
							Path: util.StringToPtr("/sysroot"),
						},
						{
							Name: "filesystem4",
							Mount: &types.Mount{
								Device:         "/dev/disk/by-partlabel/DATA2",
								Format:         "xfs",
								WipeFilesystem: true,
								Label:          util.StringToPtr("data2"),
								UUID:           util.StringToPtr("a51034e6-26b3-48df-beed-220562ac7ad1"),
								Options:        []string{"i'm an option", "me too"},
							},
						},
						{
							Name: "filesystem5",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA3",
								Format: "vfat",
							},
						},
						{
							Name: "filesystem6",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA4",
								Format: "swap",
							},
						},
					},
					Files: []types.File{
						{
							Filesystem: "filesystem1",
							Path:       "/opt/file1",
							User:       types.FileUser{Id: util.IntToPtr(500)},
							Group:      types.FileGroup{Id: util.IntToPtr(501)},
							Contents: types.FileContents{
								Inline: "file1",
							},
							Mode: 0644,
						},
						{
							Filesystem: "filesystem1",
							Path:       "/opt/file2",
							User:       types.FileUser{Id: util.IntToPtr(502)},
							Group:      types.FileGroup{Id: util.IntToPtr(503)},
							Contents: types.FileContents{
								Remote: types.Remote{
									Url:         "http://example.com/file2",
									Compression: "gzip",
									Verification: types.Verification{
										Hash: types.Hash{
											Function: "sha512",
											Sum:      "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
										},
									},
								},
							},
							Mode: 0644,
						},
						{
							Filesystem: "filesystem2",
							Path:       "/opt/file3",
							User:       types.FileUser{Id: util.IntToPtr(1000)},
							Group:      types.FileGroup{Id: util.IntToPtr(1001)},
							Contents: types.FileContents{
								Remote: types.Remote{
									Url:         "http://example.com/file3",
									Compression: "gzip",
								},
							},
							Mode: 0400,
						},
						{
							Filesystem: "filesystem2",
							Path:       "/opt/file4",
							Contents: types.FileContents{
								Inline: "",
							},
						},
					},
					Directories: []types.Directory{
						{
							Filesystem: "filesystem1",
							Path:       "/opt/dir1",
							User: types.FileUser{
								Name: "core",
							},
							Group: types.FileGroup{
								Name: "core",
							},
							Mode: 0755,
						},
					},
					Links: []types.Link{
						{
							Filesystem: "filesystem1",
							Path:       "/opt/link1",
							User: types.FileUser{
								Name: "noone",
							},
							Group: types.FileGroup{
								Name: "systemd-journald",
							},
							Target: "/opt/file2",
							Hard:   false,
						},
						{
							Filesystem: "filesystem2",
							Path:       "/opt/link2",
							Target:     "/opt/file3",
							Hard:       true,
						},
					},
				},
			}},
		},

		// systemd
		{
			in: in{data: `
systemd:
  units:
    - name: test1.service
      enable: true
      contents: test1 contents
      dropins:
        - name: conf1.conf
          contents: conf1 contents
        - name: conf2.conf
          contents: conf2 contents
    - name: test2.service
      mask: true
      contents: test2 contents
    - name: test3.service
      enabled: false
`},
			out: out{cfg: types.Config{
				Systemd: types.Systemd{
					Units: []types.SystemdUnit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []types.SystemdUnitDropIn{
								{
									Name:     "conf1.conf",
									Contents: "conf1 contents",
								},
								{
									Name:     "conf2.conf",
									Contents: "conf2 contents",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "test2 contents",
						},
						{
							Name:    "test3.service",
							Enabled: util.BoolToPtr(false),
						},
					},
				},
			}},
		},

		// networkd
		{
			in: in{data: `
networkd:
  units:
    - name: empty.netdev
    - name: test.network
      contents: test config
`},
			out: out{cfg: types.Config{
				Networkd: types.Networkd{
					Units: []types.NetworkdUnit{
						{
							Name: "empty.netdev",
						},
						{
							Name:     "test.network",
							Contents: "test config",
						},
					},
				},
			}},
		},

		// passwd
		{
			in: in{data: `
passwd:
  users:
    - name: user 1
      password_hash: password 1
      ssh_authorized_keys:
        - key1
        - key2
    - name: user 2
      password_hash: password 2
      ssh_authorized_keys:
        - key3
        - key4
      create:
        uid: 123
        gecos: gecos
        home_dir: /home/user 2
        no_create_home: true
        primary_group: wheel
        groups:
          - wheel
          - plugdev
        no_user_group: true
        system: true
        no_log_init: true
        shell: /bin/zsh
    - name: user 3
      password_hash: password 3
      ssh_authorized_keys:
        - key5
        - key6
      create: {}
    - name: user 4
      password_hash: password 4
      ssh_authorized_keys:
        - key7
        - key8
      uid: 456
      gecos: gecos
      home_dir: /home/user 4
      no_create_home: true
      primary_group: wheel
      groups:
        - wheel
        - plugdev
      no_user_group: true
      system: true
      no_log_init: true
      shell: /bin/tcsh
  groups:
    - name: group 1
      gid: 1000
      password_hash: password 1
      system: true
    - name: group 2
      password_hash: password 2
`},
			out: out{cfg: types.Config{
				Passwd: types.Passwd{
					Users: []types.User{
						{
							Name:              "user 1",
							PasswordHash:      util.StringToPtr("password 1"),
							SSHAuthorizedKeys: []string{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      util.StringToPtr("password 2"),
							SSHAuthorizedKeys: []string{"key3", "key4"},
							Create: &types.UserCreate{
								Uid:          func(i uint) *uint { return &i }(123),
								GECOS:        "gecos",
								Homedir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []string{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      util.StringToPtr("password 3"),
							SSHAuthorizedKeys: []string{"key5", "key6"},
							Create:            &types.UserCreate{},
						},
						{
							Name:              "user 4",
							PasswordHash:      util.StringToPtr("password 4"),
							SSHAuthorizedKeys: []string{"key7", "key8"},
							UID:               util.IntToPtr(456),
							Gecos:             "gecos",
							HomeDir:           "/home/user 4",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []string{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/tcsh",
						},
					},
					Groups: []types.Group{
						{
							Name:         "group 1",
							Gid:          func(i uint) *uint { return &i }(1000),
							PasswordHash: "password 1",
							System:       true,
						},
						{
							Name:         "group 2",
							PasswordHash: "password 2",
						},
					},
				},
			}},
		},
		{
			in: in{data: `
etcd:
    version: "3.0.15"
    discovery: "https://discovery.etcd.io/<token>"
    listen_client_urls: "http://0.0.0.0:2379,http://0.0.0.0:4001"
    max_wals: 44
`},
			out: out{cfg: types.Config{
				Etcd: &types.Etcd{
					Version: func(t types.EtcdVersion) *types.EtcdVersion { return &t }(
						types.EtcdVersion(semver.Version{
							Major: 3,
							Minor: 0,
							Patch: 15,
						})),
					Options: types.Etcd3_0{
						Discovery:        "https://discovery.etcd.io/<token>",
						ListenClientUrls: "http://0.0.0.0:2379,http://0.0.0.0:4001",
						MaxWals:          44,
					},
				},
			}},
		},
		{
			in: in{data: `
flannel:
    version: 0.6.2
    etcd_prefix: "/coreos.com/network2"
`},
			out: out{cfg: types.Config{
				Flannel: &types.Flannel{
					Version: func(t types.FlannelVersion) *types.FlannelVersion { return &t }(
						types.FlannelVersion(semver.Version{
							Major: 0,
							Minor: 6,
							Patch: 2,
						})),
					Options: types.Flannel0_6{
						EtcdPrefix: "/coreos.com/network2",
					},
				},
			}},
		},
	}

	for i, test := range tests {
		cfg, _, err := Parse([]byte(test.in.data))
		assert.Equal(t, test.out.r, err, "#%d: bad report", i)
		assert.Equal(t, test.out.cfg, cfg, "#%d: bad config", i)
	}
}
func TestConvertAs2_0(t *testing.T) {
	type in struct {
		cfg types.Config
	}
	type out struct {
		cfg ignTypes.Config
		r   report.Report
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{cfg: types.Config{}},
			out: out{cfg: ignTypes.Config{Ignition: ignTypes.Ignition{Version: "2.1.0"}}},
		},
		{
			in: in{cfg: types.Config{
				Networkd: types.Networkd{
					Units: []types.NetworkdUnit{
						{Name: "bad.blah", Contents: "[Match]\nName=en*\n[Network]\nDHCP=yes"},
					},
				},
			}},
			out: out{r: report.ReportFromError(errors.New("invalid networkd unit extension"), report.EntryError)},
		},
		{
			in: in{cfg: types.Config{
				Networkd: types.Networkd{
					Units: []types.NetworkdUnit{
						{Name: "bad.network", Contents: "[invalid"},
					},
				},
			}},
			out: out{r: report.ReportFromError(errors.New("invalid unit content: unable to find end of section"), report.EntryError)},
		},

		// Config
		{
			in: in{cfg: types.Config{
				Ignition: types.Ignition{
					Config: types.IgnitionConfig{
						Append: []types.ConfigReference{
							{
								Source: "http://example.com/test1",
								Verification: types.Verification{
									Hash: types.Hash{
										Function: "sha512",
										Sum:      "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
									},
								},
							},
							{
								Source: "http://example.com/test2",
							},
						},
						Replace: &types.ConfigReference{
							Source: "http://example.com/test3",
							Verification: types.Verification{
								Hash: types.Hash{
									Function: "sha512",
									Sum:      "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
								},
							},
						},
					},
				},
			}},
			out: out{cfg: ignTypes.Config{
				Ignition: ignTypes.Ignition{
					Version: "2.1.0",
					Config: ignTypes.IgnitionConfig{
						Append: []ignTypes.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "http",
									Host:   "example.com",
									Path:   "/test1",
								}).String(),
								Verification: ignTypes.Verification{
									Hash: util.StringToPtr("sha512-00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
								},
							},
							{
								Source: (&url.URL{
									Scheme: "http",
									Host:   "example.com",
									Path:   "/test2",
								}).String(),
							},
						},
						Replace: &ignTypes.ConfigReference{
							Source: (&url.URL{
								Scheme: "http",
								Host:   "example.com",
								Path:   "/test3",
							}).String(),
							Verification: ignTypes.Verification{
								Hash: util.StringToPtr("sha512-00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
							},
						},
					},
				},
			}},
		},

		//Timeouts
		{
			in: in{cfg: types.Config{
				Ignition: types.Ignition{
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(30),
						HTTPTotal:           util.IntToPtr(30),
					},
				},
			}},
			out: out{cfg: ignTypes.Config{
				Ignition: ignTypes.Ignition{
					Version: "2.1.0",
					Timeouts: ignTypes.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(30),
						HTTPTotal:           util.IntToPtr(30),
					},
				},
			}},
		},

		// Storage
		{
			in: in{cfg: types.Config{
				Storage: types.Storage{
					Disks: []types.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []types.Partition{
								{
									Label:    "ROOT",
									Number:   7,
									Size:     "100MB",
									Start:    "50MB",
									GUID:     "22222222-2222-2222-2222-222222222222",
									TypeGUID: "11111111-1111-1111-1111-111111111111",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     "1GB",
									Start:    "300MB",
									GUID:     "33333333-3333-3333-3333-333333333333",
									TypeGUID: "00000000-0000-0000-0000-000000000000",
								},
								{
									Label: "NOTHING",
								},
								{
									Label:    "ROOT_ON_RAID",
									Number:   13,
									TypeGUID: "raid_containing_root",
								},
								{
									Label:    "SWAP",
									Number:   14,
									TypeGUID: "swap_partition",
								},
								{
									Label:    "RAID",
									Number:   15,
									TypeGUID: "raid_partition",
								},
								{
									Label:    "LINUX_FS",
									Number:   16,
									TypeGUID: "linux_filesystem_data",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
					Arrays: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []string{"/dev/sdc", "/dev/sdd"},
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []string{"/dev/sde", "/dev/sdf", "/dev/sdg"},
							Spares:  1,
						},
					},
					Filesystems: []types.Filesystem{
						{
							Name: "filesystem1",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.Create{
									Force:   true,
									Options: []string{"-L", "ROOT"},
								},
							},
						},
						{
							Name: "filesystem2",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA",
								Format: "ext4",
							},
						},
						{
							Name: "filesystem3",
							Path: util.StringToPtr("/sysroot"),
						},
						{
							Name: "filesystem4",
							Mount: &types.Mount{
								Device:         "/dev/disk/by-partlabel/DATA2",
								Format:         "xfs",
								WipeFilesystem: true,
								Label:          util.StringToPtr("data2"),
								UUID:           util.StringToPtr("a51034e6-26b3-48df-beed-220562ac7ad1"),
								Options:        []string{"i'm an option", "me too"},
							},
						},
						{
							Name: "filesystem5",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA3",
								Format: "vfat",
							},
						},
						{
							Name: "filesystem6",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/DATA4",
								Format: "swap",
							},
						},
					},
					Files: []types.File{
						{
							Filesystem: "filesystem1",
							Path:       "/opt/file1",
							User:       types.FileUser{Id: util.IntToPtr(500)},
							Group:      types.FileGroup{Id: util.IntToPtr(501)},
							Contents: types.FileContents{
								Inline: "file1",
							},
							Mode: 0644,
						},
						{
							Filesystem: "filesystem1",
							Path:       "/opt/file2",
							User:       types.FileUser{Id: util.IntToPtr(502)},
							Group:      types.FileGroup{Id: util.IntToPtr(503)},
							Contents: types.FileContents{
								Remote: types.Remote{
									Url:         "http://example.com/file2",
									Compression: "gzip",
									Verification: types.Verification{
										Hash: types.Hash{
											Function: "sha512",
											Sum:      "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
										},
									},
								},
							},
							Mode: 0644,
						},
						{
							Filesystem: "filesystem2",
							Path:       "/opt/file3",
							User:       types.FileUser{Id: util.IntToPtr(1000)},
							Group:      types.FileGroup{Id: util.IntToPtr(1001)},
							Contents: types.FileContents{
								Remote: types.Remote{
									Url:         "http://example.com/file3",
									Compression: "gzip",
								},
							},
							Mode: 0400,
						},
						{
							Filesystem: "filesystem2",
							Path:       "/opt/file4",
							Contents: types.FileContents{
								Inline: "",
							},
						},
					},
					Directories: []types.Directory{
						{
							Filesystem: "filesystem1",
							Path:       "/opt/dir1",
							User: types.FileUser{
								Name: "core",
							},
							Group: types.FileGroup{
								Name: "core",
							},
							Mode: 0755,
						},
					},
					Links: []types.Link{
						{
							Filesystem: "filesystem1",
							Path:       "/opt/link1",
							User: types.FileUser{
								Name: "noone",
							},
							Group: types.FileGroup{
								Name: "systemd-journald",
							},
							Target: "/opt/file2",
							Hard:   false,
						},
						{
							Filesystem: "filesystem2",
							Path:       "/opt/link2",
							Target:     "/opt/file3",
							Hard:       true,
						},
					},
				},
			}},
			out: out{
				r: report.Report{
					Entries: []report.Entry{
						{
							Kind:    report.EntryWarning,
							Message: "the create object has been deprecated in favor of mount-level options",
						},
					},
				},
				cfg: ignTypes.Config{
					Ignition: ignTypes.Ignition{Version: "2.1.0"},
					Storage: ignTypes.Storage{
						Disks: []ignTypes.Disk{
							{
								Device:    "/dev/sda",
								WipeTable: true,
								Partitions: []ignTypes.Partition{
									{
										Label:    "ROOT",
										Number:   7,
										Size:     0x32000,
										Start:    0x19000,
										GUID:     "22222222-2222-2222-2222-222222222222",
										TypeGUID: "11111111-1111-1111-1111-111111111111",
									},
									{
										Label:    "DATA",
										Number:   12,
										Size:     0x200000,
										Start:    0x96000,
										GUID:     "33333333-3333-3333-3333-333333333333",
										TypeGUID: "00000000-0000-0000-0000-000000000000",
									},
									{
										Label: "NOTHING",
									},
									{
										Label:    "ROOT_ON_RAID",
										Number:   13,
										TypeGUID: "be9067b9-ea49-4f15-b4f6-f36f8c9e1818",
									},
									{
										Label:    "SWAP",
										Number:   14,
										TypeGUID: "0657fd6d-a4ab-43c4-84e5-0933c84b4f4f",
									},
									{
										Label:    "RAID",
										Number:   15,
										TypeGUID: "a19d880f-05fc-4d3b-a006-743f0f84911e",
									},
									{
										Label:    "LINUX_FS",
										Number:   16,
										TypeGUID: "0fc63daf-8483-4772-8e79-3d69d8477de4",
									},
								},
							},
							{
								Device:    "/dev/sdb",
								WipeTable: true,
							},
						},
						Raid: []ignTypes.Raid{
							{
								Name:    "fast",
								Level:   "raid0",
								Devices: []ignTypes.Device{"/dev/sdc", "/dev/sdd"},
							},
							{
								Name:    "durable",
								Level:   "raid1",
								Devices: []ignTypes.Device{"/dev/sde", "/dev/sdf", "/dev/sdg"},
								Spares:  1,
							},
						},
						Filesystems: []ignTypes.Filesystem{
							{
								Name: "filesystem1",
								Mount: &ignTypes.Mount{
									Device: "/dev/disk/by-partlabel/ROOT",
									Format: "btrfs",
									Create: &ignTypes.Create{
										Force:   true,
										Options: []ignTypes.CreateOption{"-L", "ROOT"},
									},
								},
							},
							{
								Name: "filesystem2",
								Mount: &ignTypes.Mount{
									Device: "/dev/disk/by-partlabel/DATA",
									Format: "ext4",
								},
							},
							{
								Name: "filesystem3",
								Path: util.StringToPtr("/sysroot"),
							},
							{
								Name: "filesystem4",
								Mount: &ignTypes.Mount{
									Device:         "/dev/disk/by-partlabel/DATA2",
									Format:         "xfs",
									WipeFilesystem: true,
									Label:          util.StringToPtr("data2"),
									UUID:           util.StringToPtr("a51034e6-26b3-48df-beed-220562ac7ad1"),
									Options:        []ignTypes.MountOption{"i'm an option", "me too"},
								},
							},
							{
								Name: "filesystem5",
								Mount: &ignTypes.Mount{
									Device: "/dev/disk/by-partlabel/DATA3",
									Format: "vfat",
								},
							},
							{
								Name: "filesystem6",
								Mount: &ignTypes.Mount{
									Device: "/dev/disk/by-partlabel/DATA4",
									Format: "swap",
								},
							},
						},
						Files: []ignTypes.File{
							{
								Node: ignTypes.Node{
									Filesystem: "filesystem1",
									Path:       "/opt/file1",
									User:       ignTypes.NodeUser{ID: util.IntToPtr(500)},
									Group:      ignTypes.NodeGroup{ID: util.IntToPtr(501)},
								},
								FileEmbedded1: ignTypes.FileEmbedded1{
									Contents: ignTypes.FileContents{
										Source: (&url.URL{
											Scheme: "data",
											Opaque: ",file1",
										}).String(),
									},
									Mode: 0644,
								},
							},
							{
								Node: ignTypes.Node{
									Filesystem: "filesystem1",
									Path:       "/opt/file2",
									User:       ignTypes.NodeUser{ID: util.IntToPtr(502)},
									Group:      ignTypes.NodeGroup{ID: util.IntToPtr(503)},
								},
								FileEmbedded1: ignTypes.FileEmbedded1{
									Contents: ignTypes.FileContents{
										Source: (&url.URL{
											Scheme: "http",
											Host:   "example.com",
											Path:   "/file2",
										}).String(),
										Compression: "gzip",
										Verification: ignTypes.Verification{
											Hash: util.StringToPtr("sha512-00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
										},
									},
									Mode: 0644,
								},
							},
							{
								Node: ignTypes.Node{
									Filesystem: "filesystem2",
									Path:       "/opt/file3",
									User:       ignTypes.NodeUser{ID: util.IntToPtr(1000)},
									Group:      ignTypes.NodeGroup{ID: util.IntToPtr(1001)},
								},
								FileEmbedded1: ignTypes.FileEmbedded1{
									Contents: ignTypes.FileContents{
										Source: (&url.URL{
											Scheme: "http",
											Host:   "example.com",
											Path:   "/file3",
										}).String(),
										Compression: "gzip",
									},
									Mode: 0400,
								},
							},
							{
								Node: ignTypes.Node{

									Filesystem: "filesystem2",
									Path:       "/opt/file4",
								},
								FileEmbedded1: ignTypes.FileEmbedded1{
									Contents: ignTypes.FileContents{
										Source: (&url.URL{
											Scheme: "data",
											Opaque: ",",
										}).String(),
									},
								},
							},
						},
						Directories: []ignTypes.Directory{
							{
								Node: ignTypes.Node{
									Filesystem: "filesystem1",
									Path:       "/opt/dir1",
									User: ignTypes.NodeUser{
										Name: "core",
									},
									Group: ignTypes.NodeGroup{
										Name: "core",
									},
								},
								DirectoryEmbedded1: ignTypes.DirectoryEmbedded1{
									Mode: 0755,
								},
							},
						},
						Links: []ignTypes.Link{
							{
								Node: ignTypes.Node{
									Filesystem: "filesystem1",
									Path:       "/opt/link1",
									User: ignTypes.NodeUser{
										Name: "noone",
									},
									Group: ignTypes.NodeGroup{
										Name: "systemd-journald",
									},
								},
								LinkEmbedded1: ignTypes.LinkEmbedded1{
									Target: "/opt/file2",
									Hard:   false,
								},
							},
							{
								Node: ignTypes.Node{
									Filesystem: "filesystem2",
									Path:       "/opt/link2",
								},
								LinkEmbedded1: ignTypes.LinkEmbedded1{
									Target: "/opt/file3",
									Hard:   true,
								},
							},
						},
					},
				},
			},
		},

		// systemd
		{
			in: in{cfg: types.Config{
				Systemd: types.Systemd{
					Units: []types.SystemdUnit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "[Service]\nType=oneshot\nExecStart=/usr/bin/echo test 1\n\n[Install]\nWantedBy=multi-user.target\n",
							Dropins: []types.SystemdUnitDropIn{
								{
									Name:     "conf1.conf",
									Contents: "[Service]\nExecStart=",
								},
								{
									Name:     "conf2.conf",
									Contents: "[Service]\nExecStart=",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "[Service]\nType=oneshot\nExecStart=/usr/bin/echo test 2\n\n[Install]\nWantedBy=multi-user.target\n",
						},
					},
				},
			}},
			out: out{cfg: ignTypes.Config{
				Ignition: ignTypes.Ignition{Version: "2.1.0"},
				Systemd: ignTypes.Systemd{
					Units: []ignTypes.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "[Service]\nType=oneshot\nExecStart=/usr/bin/echo test 1\n\n[Install]\nWantedBy=multi-user.target\n",
							Dropins: []ignTypes.Dropin{
								{
									Name:     "conf1.conf",
									Contents: "[Service]\nExecStart=",
								},
								{
									Name:     "conf2.conf",
									Contents: "[Service]\nExecStart=",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "[Service]\nType=oneshot\nExecStart=/usr/bin/echo test 2\n\n[Install]\nWantedBy=multi-user.target\n",
						},
					},
				},
			}},
		},

		// networkd
		{
			in: in{cfg: types.Config{
				Networkd: types.Networkd{
					Units: []types.NetworkdUnit{
						{
							Name:     "test.network",
							Contents: "[Match]\nName=en*\n[Network]\nDHCP=yes",
						},
						{
							Name: "empty.netdev",
						},
					},
				},
			}},
			out: out{cfg: ignTypes.Config{
				Ignition: ignTypes.Ignition{Version: "2.1.0"},
				Networkd: ignTypes.Networkd{
					Units: []ignTypes.Networkdunit{
						{
							Name:     "test.network",
							Contents: "[Match]\nName=en*\n[Network]\nDHCP=yes",
						},
						{
							Name: "empty.netdev",
						},
					},
				},
			}},
		},

		// passwd
		{
			in: in{cfg: types.Config{
				Passwd: types.Passwd{
					Users: []types.User{
						{
							Name:              "user 1",
							PasswordHash:      util.StringToPtr("password 1"),
							SSHAuthorizedKeys: []string{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      util.StringToPtr("password 2"),
							SSHAuthorizedKeys: []string{"key3", "key4"},
							Create: &types.UserCreate{
								Uid:          func(i uint) *uint { return &i }(123),
								GECOS:        "gecos",
								Homedir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []string{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      util.StringToPtr("password 3"),
							SSHAuthorizedKeys: []string{"key5", "key6"},
							Create:            &types.UserCreate{},
						},
						{
							Name:              "user 4",
							PasswordHash:      util.StringToPtr("password 4"),
							SSHAuthorizedKeys: []string{"key7", "key8"},
							UID:               util.IntToPtr(456),
							Gecos:             "gecos",
							HomeDir:           "/home/user 4",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []string{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/tcsh",
						},
					},
					Groups: []types.Group{
						{
							Name:         "group 1",
							Gid:          func(i uint) *uint { return &i }(1000),
							PasswordHash: "password 1",
							System:       true,
						},
						{
							Name:         "group 2",
							PasswordHash: "password 2",
						},
					},
				},
			}},
			out: out{
				r: report.Report{
					Entries: []report.Entry{
						{
							Kind:    report.EntryWarning,
							Message: "the create object has been deprecated in favor of user-level options",
						},
						{
							Kind:    report.EntryWarning,
							Message: "the create object has been deprecated in favor of user-level options",
						},
					},
				},
				cfg: ignTypes.Config{
					Ignition: ignTypes.Ignition{Version: "2.1.0"},
					Passwd: ignTypes.Passwd{
						Users: []ignTypes.PasswdUser{
							{
								Name:              "user 1",
								PasswordHash:      util.StringToPtr("password 1"),
								SSHAuthorizedKeys: []ignTypes.SSHAuthorizedKey{"key1", "key2"},
								Create:            nil,
							},
							{
								Name:              "user 2",
								PasswordHash:      util.StringToPtr("password 2"),
								SSHAuthorizedKeys: []ignTypes.SSHAuthorizedKey{"key3", "key4"},
								Create: &ignTypes.Usercreate{
									UID:          util.IntToPtr(123),
									Gecos:        "gecos",
									HomeDir:      "/home/user 2",
									NoCreateHome: true,
									PrimaryGroup: "wheel",
									Groups:       []ignTypes.UsercreateGroup{"wheel", "plugdev"},
									NoUserGroup:  true,
									System:       true,
									NoLogInit:    true,
									Shell:        "/bin/zsh",
								},
							},
							{
								Name:              "user 3",
								PasswordHash:      util.StringToPtr("password 3"),
								SSHAuthorizedKeys: []ignTypes.SSHAuthorizedKey{"key5", "key6"},
								Create:            &ignTypes.Usercreate{},
							},
							{
								Name:              "user 4",
								PasswordHash:      util.StringToPtr("password 4"),
								SSHAuthorizedKeys: []ignTypes.SSHAuthorizedKey{"key7", "key8"},
								UID:               util.IntToPtr(456),
								Gecos:             "gecos",
								HomeDir:           "/home/user 4",
								NoCreateHome:      true,
								PrimaryGroup:      "wheel",
								Groups:            []ignTypes.PasswdUserGroup{"wheel", "plugdev"},
								NoUserGroup:       true,
								System:            true,
								NoLogInit:         true,
								Shell:             "/bin/tcsh",
							},
						},
						Groups: []ignTypes.PasswdGroup{
							{
								Name:         "group 1",
								Gid:          util.IntToPtr(1000),
								PasswordHash: "password 1",
								System:       true,
							},
							{
								Name:         "group 2",
								PasswordHash: "password 2",
							},
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		cfg, r := ConvertAs2_0(test.in.cfg, "", nil)
		assert.Equal(t, test.out.r, r, "#%d: bad report", i)
		assert.Equal(t, test.out.cfg, cfg, "#%d: bad config", i)
	}
}

func TestAndConvertAs2_0_0(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		cfg ignTypes.Config
		r   report.Report
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in: in{data: `
networkd:
  units:
    - name: bad.blah
      contents: not valid
`},
			out: out{
				cfg: ignTypes.Config{},
				r: report.Report{Entries: []report.Entry{{
					Message: "invalid networkd unit extension",
					Kind:    report.EntryError,
					Line:    4,
					Column:  7,
				}}},
			},
		},
		{
			in: in{data: `
networkd:
  units:
    - name: bad.network
      contents: "[not valid"
`},
			out: out{
				cfg: ignTypes.Config{},
				r: report.Report{Entries: []report.Entry{{
					Message: "invalid unit content: unable to find end of section",
					Kind:    report.EntryError,
					Line:    4,
					Column:  7,
				}}},
			},
		},

		// valid
		{
			in: in{data: `
ignition:
  config:
    append:
      - source: http://example.com/test1
        verification:
          hash:
            function: sha512
            sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
      - source: http://example.com/test2
    replace:
      source: http://example.com/test3
      verification:
        hash:
          function: sha512
          sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
`},
			out: out{cfg: ignTypes.Config{
				Ignition: ignTypes.Ignition{
					Version: "2.1.0",
					Config: ignTypes.IgnitionConfig{
						Append: []ignTypes.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "http",
									Host:   "example.com",
									Path:   "/test1",
								}).String(),
								Verification: ignTypes.Verification{
									Hash: util.StringToPtr("sha512-00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
								},
							},
							{
								Source: (&url.URL{
									Scheme: "http",
									Host:   "example.com",
									Path:   "/test2",
								}).String(),
							},
						},
						Replace: &ignTypes.ConfigReference{
							Source: (&url.URL{
								Scheme: "http",
								Host:   "example.com",
								Path:   "/test3",
							}).String(),
							Verification: ignTypes.Verification{
								Hash: util.StringToPtr("sha512-00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
							},
						},
					},
				},
			}},
		},

		// Invalid files
		{
			in: in{data: `
storage:
  files:
    - path: opt/file1
      filesystem: root
      contents:
        inline: file1
      mode: 0644
      user:
        id: 500
      group:
        id: 501
    - path: /opt/file2
      filesystem: root
      contents:
        remote:
          url: httpz://example.com/file2
          compression: gzip
          verification:
            hash:
              function: sha512
              sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
      mode: 0644
      user:
        id: 502
      group:
        id: 503
`},
			out: out{
				cfg: ignTypes.Config{},
				r: report.Report{Entries: []report.Entry{
					{
						Message: "path not absolute",
						Kind:    report.EntryError,
						Line:    4,
						Column:  13,
					},
					{
						Message: "invalid url \"httpz://example.com/file2\": invalid url scheme",
						Kind:    report.EntryError,
						Line:    17,
						Column:  16,
					},
				}},
			},
		},

		// Invalid disk dimensions
		{
			in: in{data: `
storage:
  disks:
    - device: /dev/sda
      wipe_table: true
      partitions:
        - label: ROOT
          number: 7
          size: -100MB
          start: 50MB
          type_guid: 11111111-1111-1111-1111-111111111111
        - label: DATA
          number: 12
          size: 1GB
          start: -300MB
          type_guid: 00000000-0000-0000-0000-000000000000
`},
			out: out{
				cfg: ignTypes.Config{},
				r: report.Report{Entries: []report.Entry{
					{
						Message: "invalid dimension (negative): \"-100MB\"",
						Kind:    report.EntryError,
						Line:    9,
						Column:  17,
					},
					{
						Message: "invalid dimension (negative): \"-300MB\"",
						Kind:    report.EntryError,
						Line:    15,
						Column:  18,
					},
				}},
			},
		},

		// Valid files
		{
			in: in{data: `
storage:
  files:
    - path: /opt/file1
      filesystem: root
      contents:
        inline: file1
      mode: 0644
      user:
        id: 500
      group:
        id: 501
    - path: /opt/file2
      filesystem: root
      contents:
        remote:
          url: http://example.com/file2
          compression: gzip
          verification:
            hash:
              function: sha512
              sum: 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
      mode: 0644
      user:
        id: 502
      group:
        id: 503
`},
			out: out{cfg: ignTypes.Config{
				Ignition: ignTypes.Ignition{Version: "2.1.0"},
				Storage: ignTypes.Storage{
					Files: []ignTypes.File{
						{
							Node: ignTypes.Node{
								Filesystem: "root",
								Path:       "/opt/file1",
								User:       ignTypes.NodeUser{ID: util.IntToPtr(500)},
								Group:      ignTypes.NodeGroup{ID: util.IntToPtr(501)},
							},
							FileEmbedded1: ignTypes.FileEmbedded1{
								Contents: ignTypes.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
								},
								Mode: 0644,
							},
						},
						{
							Node: ignTypes.Node{
								Filesystem: "root",
								Path:       "/opt/file2",
								User:       ignTypes.NodeUser{ID: util.IntToPtr(502)},
								Group:      ignTypes.NodeGroup{ID: util.IntToPtr(503)},
							},
							FileEmbedded1: ignTypes.FileEmbedded1{
								Contents: ignTypes.FileContents{
									Source: (&url.URL{
										Scheme: "http",
										Host:   "example.com",
										Path:   "/file2",
									}).String(),
									Compression: "gzip",
									Verification: ignTypes.Verification{
										Hash: util.StringToPtr("sha512-00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
									},
								},
								Mode: 0644,
							},
						},
					},
				},
			}},
		},
	}

	for i, test := range tests {
		cfg, ast, r := Parse([]byte(test.in.data))
		if len(r.Entries) != 0 {
			t.Errorf("#%d: got error while parsing input: %v", i, r)
		}
		igncfg, r := ConvertAs2_0(cfg, "", ast)
		assert.Equal(t, test.out.r, r, "#%d: bad report", i)
		assert.Equal(t, test.out.cfg, igncfg, "#%d: bad config", i)
	}

}
