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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/golang/glog"
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
	rcd    *exec.Cmd
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

	d := &driver{
		config: cfg,
	}
	return d, d.startRCD()
}

func (d *driver) Run() error {
	s := NewNonBlockingGRPCServer()
	// hp itself implements ControllerServer, NodeServer, and IdentityServer.
	s.Start(d.config.Endpoint, d, d, d)
	s.Wait()

	var err error
	ch := make(chan struct{})
	go func() {
		err = d.coreQuit()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		glog.Infof("killing rcd %+v", d.rcd.Process.Kill())
	}
	glog.Infof("waiting tcd %+v", d.rcd.Wait())
	return err
}

func (d *driver) startRCD() error {
	args := []string{}
	if d.config.RcloneConfig != "" {
		args = append(args, "--config", d.config.RcloneConfig)
	}
	d.rcd = exec.Command("rclone", "rcd", "--rc-no-auth", "--log-level=INFO")
	d.rcd.Stdout = os.Stdout
	d.rcd.Stderr = os.Stderr
	return d.rcd.Start()
}

func (d *driver) rc(method string, data map[string]any) (gjson.Result, error) {
	var res gjson.Result

	b, err := json.Marshal(data)
	if err != nil {
		return res, err
	}
	resp, err := http.DefaultClient.Post(fmt.Sprintf("http://localhost:5572/%s", method), "application/json", bytes.NewReader(b))
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%d: %s", resp.StatusCode, all)
	} else {
		res = gjson.ParseBytes(all)
		if res.Get("error").String() != "" {
			err = fmt.Errorf("%d: %s", resp.StatusCode, all)
		}
	}
	return res, err
}

func (d *driver) remoteList() ([]string, error) {
	res, err := d.rc("config/listremotes", nil)
	v := []string{}
	for _, e := range res.Get("remotes").Array() {
		v = append(v, e.String())
	}
	return v, err
}

func (d *driver) remoteAbout(remote, path string) (gjson.Result, error) {
	return d.rc("operations/about", map[string]any{"fs": fmt.Sprintf("%s:%s", remote, path)})
}

func (d *driver) remoteCreate(remote string, parameters string) (gjson.Result, error) {
	return d.rc("config/create", map[string]any{
		"name":       remote,
		"type":       gjson.Parse(parameters).Get("type").String(),
		"parameters": parameters,
		"opt":        "{\"nonInteractive\": true}",
	})
}

func (d *driver) remoteMount(remote, rpath, target string, vfs, mount map[string]any) (res gjson.Result, err error) {
	if _, e := os.Stat(target); e != nil && errors.Is(e, os.ErrNotExist) {
		if err = os.MkdirAll(target, 0755); err != nil {
			return
		}
	}
	if _, err = os.Stat(target); err != nil {
		return
	}
	res, err = d.remoteUmount(target)
	if err != nil {
		return res, err
	}
	vb, err := json.Marshal(vfs)
	if err != nil {
		return gjson.Result{}, err
	}
	mb, err := json.Marshal(mount)
	if err != nil {
		return gjson.Result{}, err
	}
	return d.rc("mount/mount", map[string]any{
		"fs":         fmt.Sprintf("%s:%s", remote, rpath),
		"mountPoint": target,
		"mountOpt":   string(mb),
		"vfsOpt":     string(vb),
	})
}

func (d *driver) remoteUmount(target string) (res gjson.Result, err error) {
	if target == "" {
		res, err = d.rc("mount/unmountall", map[string]any{})
	} else {
		if _, e := os.Stat(target); e != nil {
			return
		}
		if e := exec.Command("mountpoint", target).Run(); e != nil {
			return
		}
		res, err = d.rc("mount/unmount", map[string]any{"mountPoint": target})
	}
	return
}

func (d *driver) coreQuit() error {
	_, err := d.rc("core/quit", nil)
	return err
}
