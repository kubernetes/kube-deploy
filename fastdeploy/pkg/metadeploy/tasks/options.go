package tasks

import (
	"encoding/json"
	"fmt"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
)

type OptionsTask struct {
	Name     string
	contents string
	Options  execution.Options
}

func (o *OptionsTask) SortKey() (int, string) {
	return TaskOrderOptions, o.Name
}

func (o *OptionsTask) String() string {
	return fmt.Sprintf("Options: %s", o.Name)
}

func NewOptionsTask(name string, contents string) (*OptionsTask, error) {
	o := &OptionsTask{Name: name, contents: contents}
	options, err := o.parse()
	if err != nil {
		return nil, err
	}
	o.Options = options
	return o, nil
}

func (o *OptionsTask) parse() (execution.Options, error) {
	var options execution.Options
	err := json.Unmarshal([]byte(o.contents), &options)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON options: %v", err)
	}

	return options, nil
}

func (o *OptionsTask) Run(c *execution.Context) error {
	err := c.MergeOptions(o.Options)
	return err
}
