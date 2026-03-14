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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/zalando/go-keyring/internal/shellescape"
)

const (
	execPathKeychain = "/usr/bin/security"

	// encodingPrefix is a well-known prefix added to strings encoded by Set.
	encodingPrefix       = "go-keyring-encoded:"
	base64EncodingPrefix = "go-keyring-base64:"
)

type macOSXKeychain struct{}

// func (*MacOSXKeychain) IsAvailable() bool {
// 	return exec.Command(execPathKeychain).Run() != exec.ErrNotFound
// }

// Get password from macos keyring given service and user name.
func (k macOSXKeychain) Get(service, username string) (string, error) {
	out, err := exec.Command(
		execPathKeychain,
		"find-generic-password",
		"-s", service,
		"-wa", username).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "could not be found") {
			err = ErrNotFound
		}
		return "", err
	}

	trimStr := strings.TrimSpace(string(out[:]))
	// if the string has the well-known prefix, assume it's encoded
	if strings.HasPrefix(trimStr, encodingPrefix) {
		dec, err := hex.DecodeString(trimStr[len(encodingPrefix):])
		return string(dec), err
	} else if strings.HasPrefix(trimStr, base64EncodingPrefix) {
		dec, err := base64.StdEncoding.DecodeString(trimStr[len(base64EncodingPrefix):])
		return string(dec), err
	}

	return trimStr, nil
}

// Set stores a secret in the macos keyring given a service name and a user.
func (k macOSXKeychain) Set(service, username, password string) error {
	// if the added secret has multiple lines or some non ascii,
	// osx will hex encode it on return. To avoid getting garbage, we
	// encode all passwords
	password = base64EncodingPrefix + base64.StdEncoding.EncodeToString([]byte(password))

	cmd := exec.Command(execPathKeychain, "-i")
	stdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	command := fmt.Sprintf("add-generic-password -U -s %s -a %s -w %s\n", shellescape.Quote(service), shellescape.Quote(username), shellescape.Quote(password))
	if len(command) > 4096 {
		return ErrSetDataTooBig
	}

	if _, err := io.WriteString(stdIn, command); err != nil {
		return err
	}

	if err = stdIn.Close(); err != nil {
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
	if strings.Contains(string(out), "could not be found") {
		err = ErrNotFound
	}
	return err
}

// DeleteAll deletes all secrets for a given service
func (k macOSXKeychain) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrNotFound
	}
	// Delete each secret in a while loop until there is no more left
	// under the service
	for {
		out, err := exec.Command(
			execPathKeychain,
			"delete-generic-password",
			"-s", service).CombinedOutput()
		if strings.Contains(string(out), "could not be found") {
			return nil
		} else if err != nil {
			return err
		}
	}

}

// ListUsers returns a list of all users for a given service
func (k macOSXKeychain) ListUsers(service string) ([]string, error) {
	if service == "" {
		return []string{}, nil
	}

	// Get the default keychain since there can be multiple keychains
	defaultKeychainOut, err := exec.Command(execPathKeychain, "default-keychain").CombinedOutput()
	if err != nil {
		return nil, err
	}
	// Take first line in case multiple default keychains are returned
	firstLine := strings.SplitN(strings.TrimSpace(string(defaultKeychainOut)), "\n", 2)[0]
	defaultKeychain := strings.Trim(strings.TrimSpace(firstLine), `"`)

	// dump-keychain (not dump-keyring) requires a keychain path
	out, err := exec.Command(execPathKeychain, "dump-keychain", defaultKeychain).CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Parse dump-keychain output. Format:
	// keychain: "/Users/username/Library/Keychains/login.keychain-db"
	//     class: "genp"
	//     attributes:
	//         "svce"<blob>="service-name"
	//         "acct"<blob>="account-name"
	valueOf := func(s, prefix string) string {
		return strings.Trim(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), prefix)), `"`)
	}

	var users []string
	seenUsers := make(map[string]bool)
	var currentKeychain, currentSvc, currentAcct string
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "keychain:"):
			// Save previous entry if it matched
			if currentKeychain == defaultKeychain && currentSvc == service && currentAcct != "" && !seenUsers[currentAcct] {
				seenUsers[currentAcct] = true
				users = append(users, currentAcct)
			}
			currentKeychain = valueOf(line, "keychain:")
			currentSvc = ""
			currentAcct = ""
		case strings.Contains(line, `"svce"`):
			if idx := strings.Index(line, `="`); idx != -1 {
				start := idx + 2
				if end := strings.Index(line[start:], `"`); end != -1 {
					currentSvc = line[start : start+end]
				}
			}
		case strings.Contains(line, `"acct"`):
			if idx := strings.Index(line, `="`); idx != -1 {
				start := idx + 2
				if end := strings.Index(line[start:], `"`); end != -1 {
					currentAcct = line[start : start+end]
				}
			}
		}
	}
	// Don't forget the last entry
	if currentKeychain == defaultKeychain && currentSvc == service && currentAcct != "" && !seenUsers[currentAcct] {
		users = append(users, currentAcct)
	}

	return users, nil
}

func init() {
	p := macOSXKeychain{}
	provider = p
	restoreProvider = func() { provider = p }
}
