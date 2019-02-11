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
	"io"
	"os/exec"
	"strings"
)

const (
	execPathKeychain = "/usr/bin/security"
)

type macOSXKeychain struct{}

// func (*MacOSXKeychain) IsAvailable() bool {
// 	return exec.Command(execPathKeychain).Run() != exec.ErrNotFound
// }

// Get gets a secret from the keyring given a service name and a user.
func (k macOSXKeychain) Get(service, username string) (string, error) {
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

// Set stores stores user and pass in the keyring under the defined service
// name.
func (k macOSXKeychain) Set(service, username, password string) error {
	cmd := exec.Command(
		execPathKeychain,
		"-i", // start in interactive mode to be able to pass the password using stdIn.
	)

	stdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	// Write the command to stdIn.
	command := fmt.Sprintf("add-generic-password -U -s %s -a %s -w %s\n", service, username, password)
	io.WriteString(stdIn, command)

	err = stdIn.Close()
	if err != nil {
		return err
	}

	err = cmd.Wait()

	return err
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k macOSXKeychain) Delete(service, username string) error {
	out, err := exec.Command(
		execPathKeychain,
		"delete-generic-password",
		"-s", service,
		"-a", username).CombinedOutput()
	if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
		err = ErrNotFound
	}
	return err
}

func init() {
	provider = macOSXKeychain{}
}
