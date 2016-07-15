// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyring

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	execPathKeychain = "/usr/bin/security"
)

type MacOSXKeychain struct{}

// func (*MacOSXKeychain) IsAvailable() bool {
// 	return exec.Command(execPathKeychain).Run() != exec.ErrNotFound
// }

func (k MacOSXKeychain) Get(service, username string) (string, error) {
	out, err := exec.Command(
		execPathKeychain,
		"find-generic-password",
		"-s", service,
		"-wa", username).CombinedOutput()
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
			err = ErrNotFound
		}
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf("%s", out)), nil
}

func (k MacOSXKeychain) Set(service, username, password string) error {
	return exec.Command(
		execPathKeychain,
		"add-generic-password",
		"-U", //update if exists
		"-s", service,
		"-a", username,
		"-w", password).Run()
}

// func (k MacOSXKeychain) Delete(service, username string) error {
// 	out, err := exec.Command(
// 		execPathKeychain,
// 		"delete-generic-password",
// 		"-s", service,
// 		"-a", username).CombinedOutput()
// 	if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
// 		err = ErrNotFound
// 	}
// 	return err
// }

func init() {
	provider = MacOSXKeychain{}
}
