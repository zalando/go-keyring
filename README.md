# Go keyring library
[![Go Report Card](https://goreportcard.com/badge/zalando/go-keyring)](https://goreportcard.com/report/zalando/go-keyring)
[![GoDoc](https://godoc.org/github.com/zalando/go-keyring?status.svg)](https://godoc.org/github.com/zalando/go-keyring)

`go-keyring` is an OS agnostic library for *setting*, *getting* and *deleting*
secrets from the system keyring. It currently support **OS X** and **Linux
(dbus)**. A NO-OP implementation for Windows is also included to make it
portable.

## Dependencies

#### OS X

The OS X implementation depends on the `/usr/bin/security` binary for
interfacing with the OS X keychain. It should be available by default.

#### Linux

The Linux implementation depends on the [Secret Service][SecretService] dbus
interface which is provided by `gnome-keyring`.

It's expected that the default collection `login` exists in the keyring, this
is default in most distros. If it doesn't exist you can create it through the
keyring frontend program `seahorse`.

 * Open `seahorse`
 * Go to **File > New > Password Keyring**
 * Click **Continue**
 * When asked for a name, use: **login**

## Usage

How to *set* and *get* a secret from the keyring.

```go
package main

import (
    "log"

    "github.com/zalando/go-keyring"
)

func main() {
    service := "my-app"
    user := "anon"
    password := "secret"

    // set password
    err := keyring.Set(service, user, password)
    if err != nil {
        log.Fatal(err)
    }

    // get password
    secret, err := keyring.Get(service, user)
    if err != nil {
        log.Fatal(err)
    }

    log.Println(secret)
}

```


## Tests

Running the tests is simple, just run:

```
go test
```

however they depend on your OS. E.g. if you run the tests on **Linux** it will
test the implementation in `keyring_linux.go` and similar if running the tests
in **OS X** it will test the implementation in `keyring_darwin.go`.

## Contributing/TODO

We welcome contributions from the community. Please use [CONTRIBUTING.md](CONTRIBUTING.md) as guideline. To help you get started, here are some items that we'd love help with:

- [Windows support](https://github.com/zalando/go-keyring/issues/3)
- [Travis OS X](https://github.com/zalando/go-keyring/issues/2)
- [Travis Linux](https://github.com/zalando/go-keyring/issues/1)
- The code base

Please use GitHub issues as the starting point for contributions, new ideas and/or bug reports.

## Contact

* E-Mail: team-teapot@zalando.de
* Security issues: Please send an email to [maintainers](MAINTAINERS). We'll try to get back to you within two workdays. If you don't hear back, then send an email to team-teapot@zalando.de and wait additional 5 days. We consider these as maximum for reply time.

## Contributors

Thanks to:

- <your name here>

## License

See [LICENSE](LICENSE) file.


[SecretService]: https://standards.freedesktop.org/secret-service/
