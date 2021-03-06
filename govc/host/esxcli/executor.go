/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

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

package esxcli

import (
	"errors"
	"fmt"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/xml"
)

type Executor struct {
	c    *govmomi.Client
	host *govmomi.HostSystem
	mme  *types.ReflectManagedMethodExecuter
	dtm  *types.InternalDynamicTypeManager
}

func NewExecutor(c *govmomi.Client, host *govmomi.HostSystem) (*Executor, error) {
	e := &Executor{
		c:    c,
		host: host,
	}

	{
		req := types.RetrieveManagedMethodExecuter{
			This: host.Reference(),
		}

		res, err := methods.RetrieveManagedMethodExecuter(c, &req)
		if err != nil {
			return nil, err
		}

		e.mme = res.Returnval
	}

	{
		req := types.RetrieveDynamicTypeManager{
			This: host.Reference(),
		}

		res, err := methods.RetrieveDynamicTypeManager(c, &req)
		if err != nil {
			return nil, err
		}

		e.dtm = res.Returnval
	}

	return e, nil
}

func (e *Executor) CommandInfo(c *Command) ([]CommandInfoParam, error) {
	req := types.ExecuteSoap{
		Moid:   "ha-dynamic-type-manager-local-cli-cliinfo",
		Method: "vim.CLIInfo.FetchCLIInfo",
		Argument: []types.ReflectManagedMethodExecuterSoapArgument{
			c.Argument("typeName", "vim.EsxCLI."+c.Namespace()),
		},
	}

	var info CommandInfo

	if err := e.Execute(&req, &info); err != nil {
		return nil, err
	}

	name := c.Name()
	for _, method := range info.Method {
		if method.Name == name {
			return method.Param, nil
		}
	}

	return nil, fmt.Errorf("method '%s' not found in name space '%s'", name, c.Namespace())
}

func (e *Executor) NewRequest(args []string) (*types.ExecuteSoap, error) {
	c := NewCommand(args)

	params, err := e.CommandInfo(c)
	if err != nil {
		return nil, err
	}

	sargs, err := c.Parse(params)
	if err != nil {
		return nil, err
	}

	sreq := types.ExecuteSoap{
		Moid:     c.Moid(),
		Method:   c.Method(),
		Argument: sargs,
	}

	return &sreq, nil
}

func (e *Executor) Execute(req *types.ExecuteSoap, res interface{}) error {
	req.This = e.mme.ManagedObjectReference
	req.Version = "urn:vim25/5.0"

	x, err := methods.ExecuteSoap(e.c, req)
	if err != nil {
		return err
	}

	if x.Returnval != nil {
		if x.Returnval.Fault != nil {
			return errors.New(x.Returnval.Fault.FaultMsg)
		}

		if err := xml.Unmarshal([]byte(x.Returnval.Response), res); err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) Run(args []string) (*Response, error) {
	req, err := e.NewRequest(args)
	if err != nil {
		return nil, err
	}

	res := &Response{}

	if err := e.Execute(req, res); err != nil {
		return nil, err
	}

	return res, nil
}
