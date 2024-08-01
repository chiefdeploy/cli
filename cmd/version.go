package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chiefdeploy/cli/lib"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the Chief version installed.",
	Long:  "Shows the Chief version installed.",
	Run: func(cmd *cobra.Command, args []string) {
		versionCommand()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func getCLIRemoteVersion() string {
	version, err := exec.Command("bash", "-c", "curl -fsSL https://install.chiefdeploy.com/version-cli").Output()

	if err != nil {
		fmt.Println(errorStyle.Render("Error: Unable to retrieve Chief version."))
		os.Exit(1)
	}

	return string(strings.ReplaceAll(string(version), "\n", ""))
}

func versionCommand() {

	if os.Geteuid() != 0 {
		fmt.Println(errorStyle.Render("Please run this command with sudo. `sudo chief install`"))
		os.Exit(1)
	}

	fmt.Println("Updating Chief CLI...")

	lib.PrintLogo()

	remote_cli_version := getCLIRemoteVersion()

	if string(remote_cli_version) != string(lib.GetCLIVersion()) {
		fmt.Println(installStyle.Render("Chief CLI is out of date. Please run `chief update` to update Chief CLI.\n"))
		os.Exit(0)
	}

	fmt.Println("Chief CLI Version: " + lib.GetCLIVersion())

	controller_version := getVersion()

	if controller_version != "" {
		fmt.Println("Chief Controller Version: " + controller_version)
	} else {
		fmt.Println("Chief Controller is not installed.")
	}

}
