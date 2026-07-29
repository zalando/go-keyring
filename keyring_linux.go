//go:build linux

package keyring

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"unicode/utf16"
)

// credManagerScript is a PowerShell preamble that defines a CredManager type via
// P/Invoke against advapi32.dll. It is a Go port of the approach used by
// keyring_wincred (https://github.com/ilpianista/keyring_wincred) and lets a WSL
// process read, write, and delete entries in the Windows Credential Manager
// without requiring any external PowerShell modules.
const credManagerScript = `Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
using System.Text;

public class CredManager {
    public const int CRED_TYPE_GENERIC = 1;
    public const int CRED_PERSIST_LOCAL_MACHINE = 2;

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    public struct CREDENTIAL {
        public int Flags;
        public int Type;
        public string TargetName;
        public string Comment;
        public System.Runtime.InteropServices.ComTypes.FILETIME LastWritten;
        public int CredentialBlobSize;
        public IntPtr CredentialBlob;
        public int Persist;
        public int AttributeCount;
        public IntPtr Attributes;
        public string TargetAlias;
        public string UserName;
    }

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    public static extern bool CredRead(string target, int type, int reservedFlag, out IntPtr credentialPtr);

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    public static extern bool CredWrite([In] ref CREDENTIAL userCredential, [In] uint flags);

    [DllImport("advapi32.dll", SetLastError = true)]
    public static extern bool CredDelete(string target, int type, int flags);

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    public static extern bool CredEnumerate(string filter, int flag, out int count, out IntPtr pCredentials);

    [DllImport("advapi32.dll", SetLastError = true)]
    public static extern void CredFree([In] IntPtr cred);

    public static string GetCredential(string target) {
        IntPtr credPtr;
        if (!CredRead(target, CRED_TYPE_GENERIC, 0, out credPtr)) {
            return null;
        }
        try {
            CREDENTIAL cred = (CREDENTIAL)Marshal.PtrToStructure(credPtr, typeof(CREDENTIAL));
            if (cred.CredentialBlobSize > 0) {
                byte[] passwordBytes = new byte[cred.CredentialBlobSize];
                Marshal.Copy(cred.CredentialBlob, passwordBytes, 0, cred.CredentialBlobSize);
                return Convert.ToBase64String(passwordBytes);
            }
            return "";
        } finally {
            CredFree(credPtr);
        }
    }

    public static bool SetCredential(string target, string username, string base64Password) {
        byte[] passwordBytes = Convert.FromBase64String(base64Password);

        CREDENTIAL cred = new CREDENTIAL();
        cred.Type = CRED_TYPE_GENERIC;
        cred.TargetName = target;
        cred.UserName = username;
        cred.CredentialBlobSize = passwordBytes.Length;
        cred.CredentialBlob = Marshal.AllocHGlobal(passwordBytes.Length);
        cred.Persist = CRED_PERSIST_LOCAL_MACHINE;

        try {
            Marshal.Copy(passwordBytes, 0, cred.CredentialBlob, passwordBytes.Length);
            return CredWrite(ref cred, 0);
        } finally {
            Marshal.FreeHGlobal(cred.CredentialBlob);
        }
    }

    public static bool DeleteCredential(string target) {
        return CredDelete(target, CRED_TYPE_GENERIC, 0);
    }
	
	public static string[] ListCredentials(string filter) {
		int count;
		IntPtr pCredentials;
		if (!CredEnumerate(filter, 0, out count, out pCredentials)) {
			return null;
		}
		try {
			string[] targets = new string[count];
			for (int i = 0; i < count; i++) {
				IntPtr credPtr = Marshal.ReadIntPtr(pCredentials, i * IntPtr.Size);
				CREDENTIAL cred = (CREDENTIAL)Marshal.PtrToStructure(credPtr, typeof(CREDENTIAL));
				targets[i] = cred.TargetName;
			}
			return targets;
		} finally {
			CredFree(pCredentials);
		}
	}
}
"@
`

// runPowerShell executes a PowerShell script through Windows interop and returns
// its exit code, trimmed stdout, and trimmed stderr. The script is passed via
// -EncodedCommand (UTF-16LE base64) to avoid any shell quoting issues.
func runPowerShell(script string) (int, string, string, error) {
	// Silence the progress stream so PowerShell does not emit the
	// "Preparing modules for first use." record. When stderr is redirected
	// (as it is here) that record is serialized as CLIXML and would otherwise
	// be mistaken for a real error.
	script = "$ProgressPreference = 'SilentlyContinue'\n" + script

	encoded := base64.StdEncoding.EncodeToString(utf16LEEncode(script))
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encoded)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), strings.TrimSpace(stdout.String()), cleanPowerShellStderr(stderr.String()), nil
		}
		return -1, "", "", err
	}
	return 0, strings.TrimSpace(stdout.String()), cleanPowerShellStderr(stderr.String()), nil
}

// cleanPowerShellStderr trims stderr and discards CLIXML-only output. When
// stderr is redirected, PowerShell serializes non-terminating streams (such as
// progress records) as a CLIXML document. Such output does not represent a real
// error, so it is treated as empty.
func cleanPowerShellStderr(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#< CLIXML") {
		return ""
	}
	return s
}

// escapePowerShellString escapes a value for safe inclusion inside a
// single-quoted PowerShell string literal.
func escapePowerShellString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// utf16LEEncode encodes a Go string as UTF-16 little-endian bytes.
func utf16LEEncode(s string) []byte {
	units := utf16.Encode([]rune(s))
	buf := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(buf[i*2:], u)
	}
	return buf
}

// utf16LEDecode decodes UTF-16 little-endian bytes into a Go string. A trailing
// odd byte, if any, is ignored.
func utf16LEDecode(b []byte) string {
	units := make([]uint16, len(b)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(units))
}

type wslKeychain struct{}

// Get gets a secret from the keyring given a service name and a user.
func (k wslKeychain) Get(service, username string) (string, error) {
	script := credManagerScript + fmt.Sprintf(`
$result = [CredManager]::GetCredential('%s')
if ($result -eq $null) {
    exit 1
}
Write-Output $result
`, escapePowerShellString(k.credName(service, username)))

	code, stdout, stderr, err := runPowerShell(script)
	if err != nil {
		return "", fmt.Errorf("wsl keystore: get: %w", err)
	}
	if code != 0 || stdout == "" {
		if stderr != "" {
			return "", fmt.Errorf("wsl keystore: get: powershell: %s", stderr)
		}
		// exit 1 with no error output means the credential was not found.
		return "", ErrNotFound
	}

	raw, err := base64.StdEncoding.DecodeString(stdout)
	if err != nil {
		return "", fmt.Errorf("wsl keystore: decode credential: %w", err)
	}
	return utf16LEDecode(raw), nil
}

// Set stores stores user and pass in the keyring under the defined service
// name.
func (k wslKeychain) Set(service, username, password string) error {
	// password may not exceed 2560 bytes (https://github.com/jaraco/keyring/issues/540#issuecomment-968329967)
	if len(password) > 2560 {
		return ErrSetDataTooBig
	}

	// service may not exceed 512 bytes (might need more testing)
	if len(service) >= 512 {
		return ErrSetDataTooBig
	}

	// service may not exceed 32k but problems occur before that
	// so we limit it to 30k
	if len(service) > 1024*30 {
		return ErrSetDataTooBig
	}

	encoded := base64.StdEncoding.EncodeToString(utf16LEEncode(password))
	script := credManagerScript + fmt.Sprintf(`
$result = [CredManager]::SetCredential('%s', '%s', '%s')
if (-not $result) {
    exit 1
}
`, escapePowerShellString(k.credName(service, username)), escapePowerShellString(username), encoded)

	code, _, stderr, err := runPowerShell(script)
	if err != nil {
		return fmt.Errorf("wsl keystore: set: %w", err)
	}
	if code != 0 {
		return fmt.Errorf("wsl keystore: set failed: %s", stderr)
	}
	return nil
}
func (k wslKeychain) delete(target string) error {
	script := credManagerScript + fmt.Sprintf(`
$result = [CredManager]::DeleteCredential('%s')
if (-not $result) {
    exit 1
}
`, escapePowerShellString(target))

	_, _, _, err := runPowerShell(script)
	if err != nil {
		return fmt.Errorf("wsl keystore: delete: %w", err)
	}
	// A non-zero exit here means the credential did not exist; treat as a no-op.
	return nil
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k wslKeychain) Delete(service, username string) error {
	_, err := k.Get(service, username) // check if it exists first
	if err != nil {
		return err
	}
	return k.delete(k.credName(service, username))
}

func (k wslKeychain) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrNotFound
	}

	creds, err := k.list(service)
	if err != nil {
		return err
	}

	deletedCount := 0

	for _, cred := range creds {
		err := k.delete(cred)
		if err != nil {
			return err
		}
		deletedCount++
	}
	return nil
}

// ListUsers returns a list of all users for a given service
func (k wslKeychain) ListUsers(service string) ([]string, error) {
	if service == "" {
		return []string{}, nil
	}

	creds, err := k.list(service)
	if err != nil {
		return nil, err
	}

	prefix := k.credName(service, "")
	var users []string

	for _, cred := range creds {
		username := strings.TrimPrefix(cred, prefix)
		if username != "" {
			users = append(users, username)
		}
	}
	sort.Strings(users)
	return slices.Compact(users), nil
}

func (k wslKeychain) list(service string) ([]string, error) {
	if service == "" {
		return []string{}, nil
	}
	script := credManagerScript + fmt.Sprintf(`
$result = [CredManager]::ListCredentials('%s')
if ($result -eq $null) {
    exit 1
}
Write-Output $result | ConvertTo-Json -Compress
`, escapePowerShellString(k.credName(service, "*")))

	code, stdout, stderr, err := runPowerShell(script)
	if err != nil {
		return nil, fmt.Errorf("wsl keystore: get: %w", err)
	}
	if code != 0 || stdout == "" {
		if stderr != "" {
			return nil, fmt.Errorf("wsl keystore: get: powershell: %s", stderr)
		}
		// exit 1 with no error output means the credential was not found.
		return nil, nil
	}

	jsonStr := strings.TrimSpace(stdout)
	if jsonStr == "null" {
		return nil, nil
	}

	// Windows PowerShell's ConvertTo-Json emits a bare string (not an array)
	// when the result contains a single element, so try that form first.
	if strings.HasPrefix(jsonStr, "[") {
		var targets []string
		if err := json.Unmarshal([]byte(jsonStr), &targets); err != nil {
			return nil, fmt.Errorf("wsl keystore: unmarshal credential list: %w", err)
		}
		return targets, nil
	}

	var target string
	if err := json.Unmarshal([]byte(jsonStr), &target); err != nil {
		return nil, fmt.Errorf("wsl keystore: unmarshal credential list: %w", err)
	}
	return []string{target}, nil
}

// credName combines service and username to a single string.
func (k wslKeychain) credName(service, username string) string {
	return service + ":" + username
}

// IsWSL inspects the environment and the kernel identification files used
// by both WSL 1 and WSL 2 to determine whether we are running under WSL.
func IsWSL() bool {
	// WSL_DISTRO_NAME is set by WSL 2; WSL_INTEROP is set when Windows interop
	// (the ability to launch .exe processes) is enabled in either version.
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}

	// Fall back to the kernel identification, which contains "microsoft" (and
	// often "WSL") on both WSL 1 and WSL 2 kernels.
	for _, path := range []string{"/proc/sys/kernel/osrelease", "/proc/version"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lower := strings.ToLower(string(data))
		if strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl") {
			return true
		}
	}

	return false
}

func init() {
	var p Keyring
	if IsWSL() {
		p = wslKeychain{}
	} else {
		p = secretServiceProvider{}
	}
	provider = p
	restoreProvider = func() { provider = p }
}
