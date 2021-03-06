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

package guest

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vmware/govmomi/vim25/types"
)

type AuthFlag struct {
	auth types.NamePasswordAuthentication
}

func (flag *AuthFlag) String() string {
	return fmt.Sprintf("%s:%s", flag.auth.Username, strings.Repeat("x", len(flag.auth.Password)))
}

func (flag *AuthFlag) Set(s string) error {
	c := strings.Split(s, ":")
	if len(c) > 0 {
		flag.auth.Username = c[0]
		if len(c) > 1 {
			flag.auth.Password = c[1]
		}
	}

	return nil
}

func (flag *AuthFlag) Register(f *flag.FlagSet) {
	flag.Set(os.Getenv("GOVC_GUEST_LOGIN"))
	f.Var(flag, "l", "Guest VM login credentials")
}

func (flag *AuthFlag) Process() error {
	return nil
}

func (flag *AuthFlag) Auth() types.BaseGuestAuthentication {
	return &flag.auth
}
