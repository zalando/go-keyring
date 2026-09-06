# Go Keyring library

[![Go Report Card](https://goreportcard.com/badge/github.com/zalando/go-keyring)](https://goreportcard.com/report/github.com/zalando/go-keyring)
[![GoDoc](https://godoc.org/github.com/zalando/go-keyring?status.svg)](https://godoc.org/github.com/zalando/go-keyring)

`go-keyring` is an OS-agnostic library for *setting*, *getting* and *deleting*
secrets from the system keyring. It supports **OS X**, **Linux/BSD (dbus)** and
**Windows**.

go-keyring was created after its authors searched for, but couldn't find, a better alternative. It aims to simplify
using statically linked binaries, which is cumbersome when relying on C bindings (as other keyring libraries do).

#### Potential Uses

If you're working with an application that needs to store user credentials
locally on the user's machine, go-keyring might come in handy. For instance, if you are writing a CLI for an API
that requires a username and password, you can store this information in the
keyring instead of having the user type it on every invocation.

## Dependencies

#### OS X

The OS X implementation depends on the `/usr/bin/security` binary for
interfacing with the OS X keychain. It should be available by default.

#### Linux and *BSD

The Linux and *BSD implementation depends on the [Secret Service][SecretService] dbus
interface, which is provided by [GNOME Keyring](https://wiki.gnome.org/Projects/GnomeKeyring).

It's expected that the default collection `login` exists in the keyring, because
it's the default in most distros. If it doesn't exist, you can create it through the
keyring frontend program [Seahorse](https://wiki.gnome.org/Apps/Seahorse):

* Open `seahorse`
* Go to **File > New > Password Keyring**
* Click **Continue**
* When asked for a name, use: **login**

## Example Usage

How to *set* and *get* a secret from the keyring:

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

## Direct CLI Usage

While this library provides a convenient Go API, you can also interact with the system keyring directly using OS-specific command-line tools. This can be useful for debugging, scripting, or understanding what the library does under the hood. You can use the CLI to set-up the secrets from a script and then access them from Go, or vice-versa.

### macOS

macOS uses the `security` command to interact with the Keychain.

**Set a password:**
```bash
security add-generic-password -U -s "service" -a "user" -w "password"
```

**Get a password:**
```bash
security find-generic-password -s "service" -wa "user"
```

**Delete a password:**
```bash
security delete-generic-password -s "service" -a "user"
```

Where:
- `-s` specifies the service name
- `-a` specifies the account/username
- `-w` specifies the password to store
- `-U` updates the password if it already exists
- The `w` option in `-wa` outputs only the password value

### Linux and *BSD

Linux and *BSD systems use the Secret Service API via D-Bus. The easiest way to interact with it from the command line is using `secret-tool`, which is part of libsecret.

**Install secret-tool (if not already installed):**
```bash
# Debian/Ubuntu
sudo apt-get install libsecret-tools

# Fedora/RHEL
sudo dnf install libsecret

# Arch Linux
sudo pacman -S libsecret
```

**Set a password:**
```bash
secret-tool store --label="Password for 'user' on 'service'" service "service" username "user"
# You'll be prompted to enter the password
```

Or provide the password directly:
```bash
echo -n "password" | secret-tool store --label="Password for 'user' on 'service'" service "service" username "user"
```

**Get a password:**
```bash
secret-tool lookup service "service" username "user"
```

**Delete a password:**
```bash
secret-tool clear service "service" username "user"
```

Note: The `service` and `username` are attributes used to identify the secret. The label is a human-readable description.

### Windows

Windows uses the Credential Manager, which can be accessed via `cmdkey` or PowerShell.

**Using cmdkey:**

**Set a password:**
```cmd
cmdkey /generic:"service:user" /user:"user" /pass:"password"
```

**Get a password:**

`cmdkey` doesn't support retrieving passwords directly. Use PowerShell instead:
```powershell
$cred = Get-StoredCredential -Target "service:user"
$cred.GetNetworkCredential().Password
```

Or using the Windows API via PowerShell:
```powershell
[System.Net.NetworkCredential]::new("", (Get-StoredCredential -Target "service:user").Password).Password
```

**Delete a password:**
```cmd
cmdkey /delete:"service:user"
```

**Using PowerShell with CredentialManager module:**

First, install the CredentialManager module:
```powershell
Install-Module -Name CredentialManager -Force
```

**Set a password:**
```powershell
New-StoredCredential -Target "service:user" -UserName "user" -Password "password" -Type Generic -Persist LocalMachine
```

**Get a password:**
```powershell
(Get-StoredCredential -Target "service:user").GetNetworkCredential().Password
```

**Delete a password:**
```powershell
Remove-StoredCredential -Target "service:user"
```

Note: On Windows, the library combines the service and username as `service:username` for the credential target name.

## Tests

### Running tests

Running the tests is simple:

```
go test
```

Which OS you use *does* matter. If you're using **Linux** or **BSD**, it will
test the implementation in `keyring_unix.go`. If running the tests
on **OS X**, it will test the implementation in `keyring_darwin.go`.

### Mocking

If you need to mock the keyring behavior for testing on systems without a keyring implementation you can call `MockInit()` which will replace the OS defined provider with an in-memory one.

```go
package implementation

import (
    "testing"

    "github.com/zalando/go-keyring"
)

func TestMockedSetGet(t *testing.T) {
    keyring.MockInit()
    err := keyring.Set("service", "user", "password")
    if err != nil {
        t.Fatal(err)
    }

    p, err := keyring.Get("service", "user")
    if err != nil {
        t.Fatal(err)
    }

    if p != "password" {
        t.Error("password was not the expected string")
    }

}

```

## Contributing/TODO

We welcome contributions from the community; please use [CONTRIBUTING.md](CONTRIBUTING.md) as your guidelines for getting started. Here are some items that we'd love help with:

* The code base
* Better test coverage

Please use GitHub issues as the starting point for contributions, new ideas and/or bug reports.

## Contact

* E-Mail: <team-teapot@zalando.de>
* Security issues: Please send an email to the [maintainers](MAINTAINERS), and we'll try to get back to you within two workdays. If you don't hear back, send an email to <team-teapot@zalando.de> and someone will respond within five days max.

## Contributors

Thanks to:

* [your name here]

## License

See [LICENSE](LICENSE) file.

[SecretService]: https://specifications.freedesktop.org/secret-service-spec/latest/


## 🌐 Web Resources & Interactive Index
- [ZIP ZAP](https://thequizzone.pages.dev/zip-zap.html)
- [HAPPY BLOCKS](https://studyplaying.github.io/happy-blocks.html)
- [K POP HUNTER HALLOWEEN FASHION](https://studyquests.github.io/k-pop-hunter-halloween-fashion.html)
- [CATEGORY DESTROY256](https://studyquests.github.io/category-destroy256.html)
- [MERGE FLOW](https://quizverses.github.io/merge-flow.html)
- [STEAL BRAINROT MONSTERS](https://studyplaying.github.io/steal-brainrot-monsters.html)
- [SLIDE THEM AWAY](https://studyquesthub.web.app/slide-them-away.html)
- [K POP PUZZLE HUNTERS](https://studyplaying.github.io/k-pop-puzzle-hunters.html)
- [HOME ISLAND](https://quizverses.pages.dev/home-island.html)
- [CATEGORY POOL17](https://quizverses.github.io/category-pool17.html)
- [SHOTTING BALLS](https://quizverses.github.io/shotting-balls.html)
- [CATEGORY BIKE 3](https://quizverses.github.io/category-bike-3.html)
- [CATEGORY ESCAPE](https://studyplaying.github.io/category-escape.html)
- [FAIRY WINGERELLA](https://studyquests.github.io/fairy-wingerella.html)
- [GREATSWORD V3](https://studyquests.github.io/greatsword-v3.html)
- [MERGE HEROES](https://studyquests.github.io/merge-heroes.html)
- [POP THEM](https://quizverses.pages.dev/pop-them.html)
- [TRANSFORMERS BATTLE FOR THE CITY](https://quizverses.github.io/transformers-battle-for-the-city.html)
- [CITYQUEST](https://quizverses.github.io/cityquest.html)
- [KIOMET COM](https://quizverses.pages.dev/kiomet-com.html)
- [CATEGORY SIMULATION 5](https://studyplaying.github.io/category-simulation-5.html)
- [CHALLENGER CITY DRIVER](https://studyquests.github.io/challenger-city-driver.html)
- [ITALIAN BRAINROT NEURO BEASTS](https://studyquests.github.io/italian-brainrot-neuro-beasts.html)
- [CATEGORY 2D1 070](https://quizverses.github.io/category-2d1-070.html)
- [CATEGORY TETRIS36](https://studyplaying.github.io/category-tetris36.html)
- [CATEGORY ADVENTURE 4](https://studyquests.github.io/category-adventure-4.html)
- [CATEGORY SNAKE](https://studyquests.github.io/category-snake.html)
- [WATER SORT PUZZLE 3](https://quizverses.github.io/water-sort-puzzle-3.html)
- [JUICY MATCH](https://studyquests.github.io/juicy-match.html)
- [POP PARTY SUIKA WATERMELON](https://studyquests.github.io/pop-party-suika-watermelon.html)
- [WORDMIX](https://studyquests.github.io/wordmix.html)
- [CATEGORY FLASH 3](https://quizverses.github.io/category-flash-3.html)
- [FREE THE BALL](https://studyplaying.github.io/free-the-ball.html)
- [COSMIC AVIATOR](https://studyplaying.github.io/cosmic-aviator.html)
- [DOODLE DINO RUN](https://studyquests.pages.dev/doodle-dino-run.html)
- [ATHENA MATCH 2](https://quizverses.github.io/athena-match-2.html)
- [RED STICKMAN VS CRAFTMANS 2](https://studyquests.github.io/red-stickman-vs-craftmans-2.html)
- [IDLE FIREFIGHTER 3D](https://studyquests.pages.dev/idle-firefighter-3d.html)
- [CATEGORY TOWER DEFENSE 2](https://studyplaying.github.io/category-tower-defense-2.html)
- [POPCATS MERGE THE CATS](https://studyquests.github.io/popcats-merge-the-cats.html)
- [CATEGORY MERGE GAMES](https://quizverses-9d2f2.web.app/category-merge-games.html)
- [THATS NOT MY NEIGHBOR](https://quizverses.github.io/thats-not-my-neighbor.html)
- [BRAIN TEST IQ CHALLENGE 2](https://studyquests.github.io/brain-test-iq-challenge-2.html)
- [MERGE ARCHER DEFENSE](https://quizverses.github.io/merge-archer-defense.html)
- [CATEGORY UNBLOCKED](https://studyquests.github.io/category-unblocked.html)
- [COLOR WATER PUZZLE](https://studyquests.pages.dev/color-water-puzzle.html)
- [CATEGORY SIMULATION 3](https://studyquests.github.io/category-simulation-3.html)
- [SAVE THE BEAUTY](https://quizverses-9d2f2.web.app/save-the-beauty.html)
- [CATEGORY BIKE](https://studyquests.github.io/category-bike.html)
- [CATEGORY HALLOWEEN45](https://studyquests.github.io/category-halloween45.html)
- [UPHILL RUSH 13](https://studyquests.pages.dev/uphill-rush-13.html)
- [CATEGORY RACING127](https://studyquests.github.io/category-racing127.html)
- [POOL DUEL](https://studyquests.pages.dev/pool-duel.html)
- [CATEGORY IDLE448](https://studyquests.github.io/category-idle448.html)
- [NUBIK IN THE MONSTER WORLD](https://quizverses.pages.dev/nubik-in-the-monster-world.html)
- [CATEGORY 1 PLAYER139](https://quizverses.github.io/category-1-player139.html)
- [PANDA KITCHEN IDLE TYCOON](https://studyquests.github.io/panda-kitchen-idle-tycoon.html)
- [BOMBER BATTLE ARENA](https://quizverses.pages.dev/bomber-battle-arena.html)
- [QUIZ 10 SECONDS MATH](https://studyquests.github.io/quiz-10-seconds-math.html)
- [BULL RUNNER](https://studyquests.pages.dev/bull-runner.html)
- [CATEGORY BASKETBALL](https://studyquests.github.io/category-basketball.html)
- [DREAMS](https://studyquests.github.io/dreams.html)
- [AUTUMN GLAM GALA](https://quizverses.pages.dev/autumn-glam-gala.html)
- [SUPERMARKET SIMULATOR DREAM STORE](https://studyquests.pages.dev/supermarket-simulator-dream-store.html)
- [CATEGORY CRAFTING45](https://quizverses.github.io/category-crafting45.html)
- [CAPYBARA MUKBANG ASMR](https://studyquests.github.io/capybara-mukbang-asmr.html)
- [PRIVACY](https://studyquests.github.io/privacy.html)
- [100 HIDDEN CAPYBARAS](https://studyquests.pages.dev/100-hidden-capybaras.html)
- [HAPPY TOWN](https://studyquests.github.io/happy-town.html)
- [IDLE MINER](https://studyquests.github.io/idle-miner.html)
- [GALAXY CARNAGE](https://studyquests.github.io/galaxy-carnage.html)
- [COLOR YARN SORT](https://quizverses.github.io/color-yarn-sort.html)
- [HYPERSPACE   QUANTUM FRACTURE FEZ](https://quizverses.github.io/hyperspace---quantum-fracture-fez.html)
- [PIN DETECTIVE](https://quizverses.github.io/pin-detective.html)
- [CATEGORY STRATEGY 2](https://quizverses.github.io/category-strategy-2.html)
- [IDLE HOTEL EMPIRE](https://quizverses.github.io/idle-hotel-empire.html)
- [BEAR VS HUMANS](https://thelearnquester.web.app/bear-vs-humans.html)
- [CATEGORY MMO25](https://learnquesters.pages.dev/category-mmo25.html)
- [APOCALYPSE SHELTER](https://learnquesters.pages.dev/apocalypse-shelter.html)
- [EYE ATTACK TOILET MONSTER WAR](https://learnquesters.pages.dev/eye-attack-toilet-monster-war.html)
- [DARK STONES CARD BATTLE RPG](https://studyquests.github.io/dark-stones-card-battle-rpg.html)
- [OCEAN KIDS BACK TO SCHOOL](https://quizverses.pages.dev/ocean-kids-back-to-school.html)
- [CATEGORY BUILDING182](https://thelearnquesters.pages.dev/category-building182.html)
- [CATEGORY CASUAL 12](https://thelearnquesters.pages.dev/category-casual-12.html)
- [ARCHERY LEGENDS](https://quizverses.pages.dev/archery-legends.html)
- [CATEGORY BUILDING179](https://thelearnquesters.pages.dev/category-building179.html)
- [LAST UFO DEFENSE](https://quizverses.pages.dev/last-ufo-defense.html)
- [CATEGORY WORLD CUP17](https://studyplaying.github.io/category-world-cup17.html)
- [CATEGORY ESCAPE 3](https://thelearnquesters.pages.dev/category-escape-3.html)
- [CATEGORY BASKETBALL 2](https://thelearnquester.web.app/category-basketball-2.html)
- [CATEGORY CARDS](https://quizverses.github.io/category-cards.html)
- [KICK THE NOOBIK 3D](https://quizverses.pages.dev/kick-the-noobik-3d.html)
- [JELLY MATH 3D](https://studyquests.github.io/jelly-math-3d.html)
- [NEKOS ADVENTURE](https://studyquests.github.io/nekos-adventure.html)
- [MATH STARS](https://quizverses.pages.dev/math-stars.html)
- [CATEGORY JUMPING147](https://studyquests.github.io/category-jumping147.html)
- [CATEGORY PENALTY](https://studyplayings.web.app/category-penalty.html)
- [CATEGORY ADVENTURE 2](https://quizverses.github.io/category-adventure-2.html)
- [BALL JUMP SWITCH THE COLORS](https://learnquester.github.io/ball-jump-switch-the-colors.html)
- [CATEGORY BIKE63](https://thelearnquesters.pages.dev/category-bike63.html)
- [CATEGORY SNIPER39](https://studyplayings.web.app/category-sniper39.html)
- [CATEGORY DRESS UP GAMES](https://thelearnquesters.pages.dev/category-dress-up-games.html)
- [BUTTERFLY EAR CUFF JEWELRY](https://studyquests.pages.dev/butterfly-ear-cuff-jewelry.html)
- [CATEGORY ADVENTURE 3](https://quizverses.github.io/category-adventure-3.html)
- [CATEGORY CAN T STOP PLAYING212](https://thelearnquesters.pages.dev/category-can-t-stop-playing212.html)
- [CATEGORY SCHOOL UNBLOCKER](https://studyquests.github.io/category-school-unblocker.html)
- [ONE HERO](https://learnquesters.pages.dev/one-hero.html)
- [BFFS K POP FANGIRLS](https://learnquesters.pages.dev/bffs-k-pop-fangirls.html)
- [SPRUNKI GETS SURGERY](https://studyquests.github.io/sprunki-gets-surgery.html)
- [FLYING BILL](https://studyquests.github.io/flying-bill.html)
- [CATEGORY PUZZLE 5](https://learnquester.github.io/category-puzzle-5.html)
- [UNSCREW WOOD PUZZLE](https://studyplayings.pages.dev/unscrew-wood-puzzle.html)
- [CATEGORY CAN T STOP PLAYING215](https://thelearnquesters.pages.dev/category-can-t-stop-playing215.html)
- [BATTLE RACING STARS](https://learnquesters.pages.dev/battle-racing-stars.html)
- [BANANA BOUNCE](https://thelearnquester.web.app/banana-bounce.html)
- [TAP GO DELUXE](https://quizverses.github.io/tap-go-deluxe.html)
- [CATEGORY POINT AND CLICK123](https://studyplayings.pages.dev/category-point-and-click123.html)
- [CATEGORY BIKE 3](https://thelearnquesters.pages.dev/category-bike-3.html)
- [LOGIC BLAST EXPLORER](https://learnquesters.pages.dev/logic-blast-explorer.html)
- [ROYAL PUZZLE BURST](https://studyquests.pages.dev/royal-puzzle-burst.html)
- [CATEGORY CASUAL 16](https://thelearnquesters.pages.dev/category-casual-16.html)
- [ARCHERS RANDOM](https://studyplaying.github.io/archers-random.html)
- [KOBOLM RESCUE](https://studyquests.github.io/kobolm-rescue.html)
- [CATEGORY BATTLE524](https://thelearnquesters.pages.dev/category-battle524.html)
- [FAR ORION NEW WORLDS](https://thelearnquesters.pages.dev/far-orion-new-worlds.html)
- [TILE CONNECT CLUB](https://studyplaying.github.io/tile-connect-club.html)
- [TOWER DEFENDERS](https://thelearnquesters.pages.dev/tower-defenders.html)
- [CATEGORY DRAWING34](https://thelearnquester.web.app/category-drawing34.html)
- [BLOCK SNIPER](https://studyquests.github.io/block-sniper.html)
- [AYLA WORLD PRINCESS LIFE](https://studyplaying.github.io/ayla-world-princess-life.html)
