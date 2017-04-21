package tasks

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/utils"
	"text/template"
)

type Template struct {
	Name     string
	DestPath string

	template *template.Template
	contents string

	Mode string `json:"mode"`
}

func (s *Template) SortKey() (int, string) {
	return TaskOrderTemplate, s.DestPath
}

func (t *Template) String() string {
	return fmt.Sprintf("Template: %s -> %s", t.Name, t.DestPath)
}

func NewTemplate(name string, destPath string, contents string, meta string) (*Template, error) {
	t := &Template{Name: name, DestPath: destPath}

	tmpl := template.New(name)

	funcs := template.FuncMap(map[string]interface{}{
		"base64decode": base64decode,
	})
	tmpl.Funcs(funcs)

	_, err := tmpl.Parse(contents)
	if err != nil {
		return nil, fmt.Errorf("error parsing template for %q: %v", name, err)
	}

	t.template = tmpl
	t.contents = contents

	if meta != "" {
		err := json.Unmarshal([]byte(meta), t)
		if err != nil {
			return nil, fmt.Errorf("error parsing meta for template %q: %v", name, err)
		}
	}

	return t, nil
}

func base64decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *Template) Run(c *execution.Context) error {
	var writer bytes.Buffer

	data := map[string]interface{}{}
	data["Options"] = c.Options
	data["Context"] = c
	err := t.template.Execute(&writer, data)
	if err != nil {
		return fmt.Errorf("error executing template %q: %v", t.Name, err)
	}

	fileMode, err := utils.ParseFileMode(t.Mode, 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %q: %q", t.DestPath, t.Mode)
	}

	_, err = utils.WriteFile(t.DestPath, writer.Bytes(), fileMode, 0755)
	if err != nil {
		return fmt.Errorf("error writing file %q: %v", t.DestPath, err)
	}

	return nil
}
