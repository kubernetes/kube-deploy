package execution

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
)

type Context struct {
	state   map[string]interface{}
	Options Options
	Assets  *AssetStore
	Tmpdir  string
}

func NewContext(assets *AssetStore, options Options) (*Context, error) {
	c := &Context{}
	c.state = make(map[string]interface{})
	c.Options = options
	c.Assets = assets
	t, err := ioutil.TempDir("", "deploy")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory: %v", err)
	}
	c.Tmpdir = t
	return c, nil
}

func (c *Context) Close() {
	glog.V(2).Infof("deleting temp dir: %q", c.Tmpdir)
	if c.Tmpdir != "" {
		err := os.RemoveAll(c.Tmpdir)
		if err != nil {
			glog.Warningf("unable to delete temporary directory %q: %v", c.Tmpdir, err)
		}
	}
}
func (c *Context) GetState(key string, builder func() (interface{}, error)) (interface{}, error) {
	v := c.state[key]
	if v == nil {
		var err error
		v, err = builder()
		if err != nil {
			return nil, err
		}
		c.state[key] = v
	}
	return v, nil
}

func (c *Context) MergeOptions(options Options) error {
	return c.Options.Merge(options)
}

func (c *Context) NewTempDir(prefix string) (string, error) {
	t, err := ioutil.TempDir(c.Tmpdir, prefix)
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory: %v", err)
	}
	return t, nil
}
