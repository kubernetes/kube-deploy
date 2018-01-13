/*
Copyright 2017 The Kubernetes Authors.

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
	"flag"
	"time"

	"github.com/golang/glog"
	machineclientset "k8s.io/kube-deploy/machine-api-generic-worker/pkg/client/clientset/versioned"
	machineinformers "k8s.io/kube-deploy/machine-api-generic-worker/pkg/client/informers/externalversions"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/controller"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/machines"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/signals"
	"k8s.io/kube-deploy/machine-api-generic-worker/pkg/ssh"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL   string
	kubeconfig  string
	workerCount int
)

func main() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&workerCount, "worker-count", 5, "Number of workers to process machines. Using a high number with a lot of machines might cause getting rate-limited from your cloud provider.")

	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	extclient := apiextclient.NewForConfigOrDie(cfg)
	err = machines.EnsureCustomResourceDefinitions(extclient)
	if err != nil {
		glog.Fatalf("failed to create CustomResourceDefinition: %v", err)
	}

	machineClient, err := machineclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %v", err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	machineInformerFactory := machineinformers.NewSharedInformerFactory(machineClient, time.Second*30)

	keypair, err := ssh.EnsureSSHKeypairSecret(kubeClient)
	if err != nil {
		glog.Fatalf("failed to get/create ssh keypair configmap: %v", err)
	}

	c := controller.NewMachineController(kubeClient, machineClient, kubeInformerFactory, machineInformerFactory, keypair)

	go kubeInformerFactory.Start(stopCh)
	go machineInformerFactory.Start(stopCh)

	if err = c.Run(workerCount, stopCh); err != nil {
		glog.Fatalf("Error running controller: %v", err)
	}
}
