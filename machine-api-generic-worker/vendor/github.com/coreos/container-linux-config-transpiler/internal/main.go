// Copyright 2015 CoreOS, Inc.
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/coreos/container-linux-config-transpiler/config"
	"github.com/coreos/container-linux-config-transpiler/config/platform"
	"github.com/coreos/container-linux-config-transpiler/internal/version"
)

func stderr(f string, a ...interface{}) {
	out := fmt.Sprintf(f, a...)
	fmt.Fprintln(os.Stderr, strings.TrimSuffix(out, "\n"))
}

func main() {
	flags := struct {
		help     bool
		pretty   bool
		version  bool
		inFile   string
		outFile  string
		strict   bool
		platform string
	}{}

	flag.BoolVar(&flags.help, "help", false, "Print help and exit.")
	flag.BoolVar(&flags.pretty, "pretty", false, "Indent the resulting Ignition config.")
	flag.BoolVar(&flags.version, "version", false, "Print the version and exit.")
	flag.StringVar(&flags.inFile, "in-file", "", "Path to the container linux config. Standard input unless specified otherwise.")
	flag.StringVar(&flags.outFile, "out-file", "", "Path to the resulting Ignition config. Standard output unless specified otherwise.")
	flag.BoolVar(&flags.strict, "strict", false, "Fail if any warnings are encountered.")
	flag.StringVar(&flags.platform, "platform", "", fmt.Sprintf("Platform to target. Accepted values: %v.", platform.Platforms))

	flag.Parse()

	if flags.help {
		flag.Usage()
		return
	}

	if flags.version {
		fmt.Println(version.String)
		return
	}

	var inFile *os.File
	var outFile *os.File

	if flags.inFile == "" {
		inFile = os.Stdin
	} else {
		var err error
		inFile, err = os.Open(flags.inFile)
		if err != nil {
			stderr("Failed to open: %v", err)
			os.Exit(1)
		}
	}

	if flags.outFile == "" {
		outFile = os.Stdout
	} else {
		var err error
		outFile, err = os.Create(flags.outFile)
		if err != nil {
			stderr("Failed to create: %v", err)
			os.Exit(1)
		}
	}

	dataIn, err := ioutil.ReadAll(inFile)
	if err != nil {
		stderr("Failed to read: %v", err)
		os.Exit(1)
	}

	cfg, ast, report := config.Parse(dataIn)
	if len(report.Entries) > 0 {
		stderr(report.String())
	}
	if report.IsFatal() || (flags.strict && len(report.Entries) > 0) {
		stderr("Failed to parse config")
		os.Exit(1)
	}

	ignCfg, report := config.ConvertAs2_0(cfg, flags.platform, ast)
	if len(report.Entries) > 0 {
		stderr(report.String())
		if report.IsFatal() || flags.strict {
			os.Exit(1)
		}
	}

	var dataOut []byte
	if flags.pretty {
		dataOut, err = json.MarshalIndent(&ignCfg, "", "  ")
		dataOut = append(dataOut, '\n')
	} else {
		dataOut, err = json.Marshal(&ignCfg)
	}
	if err != nil {
		stderr("Failed to marshal output: %v", err)
		os.Exit(1)
	}

	if _, err := outFile.Write(dataOut); err != nil {
		stderr("Failed to write: %v", err)
		os.Exit(1)
	}
}
