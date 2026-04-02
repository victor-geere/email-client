package auth

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// openBrowser attempts to open the given URL in the user's default browser.
// Returns an error if the platform is unsupported or the command fails.
var openBrowser = func(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}

// copyToClipboard attempts to copy text to the system clipboard.
// Silently fails if the clipboard command is unavailable.
var copyToClipboard = func(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		return
	}
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}

// promptDeviceCode displays the device code message, copies the user code to
// clipboard, and opens the verification URL in the browser.
func promptDeviceCode(dc *DeviceCodeResponse) {
	fmt.Fprintf(os.Stderr, "\n%s\n", dc.Message)

	copyToClipboard(dc.UserCode)
	fmt.Fprintf(os.Stderr, "  (Code copied to clipboard)\n")

	if err := openBrowser(dc.VerificationURI); err == nil {
		fmt.Fprintf(os.Stderr, "  Browser opened — complete sign-in and any MFA prompts there.\n\n")
	} else {
		fmt.Fprintf(os.Stderr, "  Open the URL above in your browser to sign in.\n\n")
	}
}
