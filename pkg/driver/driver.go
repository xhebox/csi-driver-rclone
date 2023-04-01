/*
MIT License

Copyright (c) 2023 xhe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package driver

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
)

type Config struct {
	NodeID        string
	Endpoint      string
	PluginName    string
	PluginVersion string
	RcloneConfig  string
}

type driver struct {
	config Config
}

func NewDriver(cfg Config) (*driver, error) {
	if cfg.NodeID == "" {
		return nil, errors.New("no node id provided")
	}

	if cfg.Endpoint == "" {
		return nil, errors.New("no driver endpoint provided")
	}

	cfg.PluginName = "csi-rclone"
	cfg.PluginVersion = "v0.1"

	return &driver{
		config: cfg,
	}, nil
}

func (d *driver) Run() error {
	s := NewNonBlockingGRPCServer()
	// hp itself implements ControllerServer, NodeServer, and IdentityServer.
	s.Start(d.config.Endpoint, d, d, d)
	s.Wait()
	return nil
}

func (d *driver) exec(arg ...string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	args := []string{}
	if d.config.RcloneConfig != "" {
		args = append(args, "--config", d.config.RcloneConfig)
	}
	args = append(args, arg...)
	cmd := exec.Command("/bin/rclone", args...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	return buf, cmd.Run()
}

func (d *driver) listremotes() ([]string, error) {
	res, err := d.exec("listremotes")
	if err != nil {
		return nil, err
	}
	str := strings.ReplaceAll(res.String(), "\n", "")
	return strings.Split(str, ":"), nil
}

func (d *driver) aboutremote(r string) (gjson.Result, error) {
	res, err := d.exec("about", r+":")
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(res.String()), nil
}
