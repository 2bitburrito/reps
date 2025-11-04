package cli

import "os/exec"

func CheckInstalledBinaries() error {
	cmd := exec.Command("gh", "version")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("fzf", "--version")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
