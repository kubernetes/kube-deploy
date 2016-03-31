package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"strings"
)

func main() {
	var flagTags string
	flag.StringVar(&flagTags, "tags", "", "tags to set")
	var flagTemplate string
	flag.StringVar(&flagTemplate, "template", "template", "template dir")

	flag.Set("logtostderr", "true")
	flag.Parse()

	tags := make(map[string]struct{})
	for _, tag := range strings.Split(flagTags, ",") {
		tags[tag] = struct{}{}
	}
	runner := &metadeploy.Runner{
		Basedir: flagTemplate,
		Tags:    tags,
	}
	err := runner.Run()
	if err != nil {
		glog.Exitf("error running template: %v", err)
	}

	tasks := runner.Tasks
	for _, task := range tasks {
		fmt.Printf("%v\n", task)
	}

	options := execution.NewOptions()

	{
		// TODO: Parse as yaml?
		// TODO: Source from metadata service or flag-passed url
		//kubeenvPath := "/var/cache/kubernetes-install/kube_env.yaml"
		kubeenvPath := "/etc/kubernetes/kube_env.yaml"
		kubeenv, err := ioutil.ReadFile(kubeenvPath)
		if err != nil {
			glog.Exitf("error reading %q", kubeenvPath)
		}

		for _, line := range strings.Split(string(kubeenv), "\n") {
			if line == "" {
				continue
			}
			tokens := strings.SplitN(line, ":", 2)
			if len(tokens) != 2 {
				glog.Exitf("unable to parse kube-env line: %q", line)
			}

			k := tokens[0]
			v := strings.TrimSpace(tokens[1])
			v = strings.Trim(v, "'")

			options[k] = v
		}
	}

	assets := execution.NewAssetStore("/var/cache/fastdeploy/assets")

	//releaseURL := "https://storage.googleapis.com/kubernetes-release/release/v1.2.0/kubernetes.tar.gz"
	//err = assets.AddURL(releaseURL, "")
	//if err != nil {
	//	glog.Exitf("error adding asset %q: %v", releaseURL, err)
	//}

	// Crazy... this is a tar.gz in the release, and is actually the one we want :-(
	// TODO: Automate extraction, or package sensibly
	serverFile := "/tmp/kubernetes-server-linux-amd64.tar.gz"
	err = assets.AddFile(serverFile)
	if err != nil {
		glog.Exitf("error adding asset %q: %v", serverFile, err)
	}

	context, err := execution.NewContext(assets, options)
	if err != nil {
		glog.Exitf("error building context: %v", err)
	}
	defer context.Close()
	for _, task := range tasks {
		fmt.Printf("Executing %v\n", task)
		err := task.Run(context)
		if err != nil {
			glog.Exitf("error running tasks (%s): %v", task, err)
		}
	}

}
