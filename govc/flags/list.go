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

package flags

import (
	"errors"
	"flag"
	"path"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/govc/flags/list"
)

type ListRelativeFunc func() (govmomi.Reference, error)

type ListFlag struct {
	*ClientFlag
	*OutputFlag
}

func (flag *ListFlag) Register(f *flag.FlagSet) {}

func (flag *ListFlag) Process() error { return nil }

func (flag *ListFlag) ListSlice(args []string, tl bool, fn ListRelativeFunc) ([]list.Element, error) {
	var out []list.Element

	for _, arg := range args {
		es, err := flag.List(arg, tl, fn)
		if err != nil {
			return nil, err
		}

		out = append(out, es...)
	}

	return out, nil
}

func (flag *ListFlag) List(arg string, tl bool, fn ListRelativeFunc) ([]list.Element, error) {
	c, err := flag.ClientFlag.Client()
	if err != nil {
		return nil, err
	}

	root := list.Element{
		Path:   "/",
		Object: c.RootFolder(),
	}

	parts := list.ToParts(arg)

	if len(parts) > 0 {
		switch parts[0] {
		case "..": // Not supported; many edge case, little value
			return nil, errors.New("cannot traverse up a tree")
		case ".": // Relative to whatever
			pivot, err := fn()
			if err != nil {
				return nil, err
			}

			mes, err := c.Ancestors(pivot)
			if err != nil {
				return nil, err
			}

			for _, me := range mes {
				// Skip root entity in building inventory path.
				if me.Parent == nil {
					continue
				}
				root.Path = path.Join(root.Path, me.Name)
			}

			root.Object = pivot
			parts = parts[1:]
		}
	}

	r := list.Recurser{
		Client:        c,
		All:           flag.JSON,
		TraverseLeafs: tl,
	}

	es, err := r.Recurse(root, parts)
	if err != nil {
		return nil, err
	}

	return es, nil
}
