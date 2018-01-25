/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"

	controllerlib "github.com/kubernetes-incubator/apiserver-builder/pkg/controller"
	"github.com/spf13/pflag"

	"k8s.io/kube-deploy/ext-apiserver/pkg/controller"
	"k8s.io/kube-deploy/ext-apiserver/pkg/controller/config"
)

func main() {

	pflag.Parse()
	config, err := controllerlib.GetConfig(config.ControllerConfig.Kubeconfig)
	if err != nil {
		log.Fatalf("Could not create Config for talking to the apiserver: %v", err)
	}

	controllers, _ := controller.GetAllControllers(config)
	controllerlib.StartControllerManager(controllers...)

	// Blockforever
	select {}
}
