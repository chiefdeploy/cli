package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/chiefdeploy/cli/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// installCmd represents the install command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the Chief controller.",
	Long:  "Updates the Chief controller.",
	Run: func(cmd *cobra.Command, args []string) {
		updateFunction(cmd)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("cron", "c", false, "Disable console output")
}

func getVersion() string {
	version := viper.GetString("current_version")

	return version
}

func updateFunction(cmd *cobra.Command) {
	var domain string
	var password string

	if os.Geteuid() != 0 {
		fmt.Println(errorStyle.Render("Please run this command with sudo. `sudo chief install`"))
		os.Exit(1)
	}

	if viper.GetString("domain") != "" {
		domain = viper.GetString("domain")
	} else {
		fmt.Println(errorStyle.Render("Error: Please run `chief install` first."))
		os.Exit(1)
	}

	if viper.GetString("password") != "" {
		password = viper.GetString("password")
	} else {
		fmt.Println(errorStyle.Render("Error: Please run `chief install` first."))
	}

	if !cmd.Flags().Changed("cron") {
		fmt.Println("Updating Chief controller...")

		lib.PrintLogo()
	}

	remote_cli_version := getCLIRemoteVersion()

	if remote_cli_version != lib.GetCLIVersion() {
		if !cmd.Flags().Changed("cron") {
			fmt.Println(installStyle.Render("Chief CLI is out of date. Please run `chief update cli` to update Chief CLI.\n"))
		}

		lib.UpdateCLI("https://chief-install.s3.eu-central-1.amazonaws.com/chief-linux-amd64")
		os.Exit(0)
	}

	err_check := exec.Command("docker", "stack", "ps", "chief").Run()

	if err_check != nil {

		fmt.Println(errorStyle.Render("Error: Chief controller is not installed. Please run `chief install` first."))
		os.Exit(1)
	}

	unformatted_version, err := exec.Command("bash", "-c", "curl -fsSL https://install.chiefdeploy.com/version").Output()

	if err != nil {
		fmt.Println(errorStyle.Render("Error: Unable to retrieve Chief version."))
		os.Exit(1)
	}

	version := strings.ReplaceAll(string(unformatted_version), "\n", "")

	if string(version) == getVersion() {
		if !cmd.Flags().Changed("cron") {
			fmt.Println(errorStyle.Render("Error: Chief is already up to date."))
		}

		os.Exit(1)
	}

	if !cmd.Flags().Changed("cron") {
		fmt.Println(installStyle.Render("Chief version: " + string(version)))
	}

	if !cmd.Flags().Changed("cron") {
		err_download_stack := spinner.New().
			Title("Downloading stack.yml...").
			Action(func() {
				err := exec.Command("bash", "-c", "export DOMAIN="+domain+"; export PASSWORD="+password+"; export CHIEF_VERSION="+string(version)+"; curl -fsSL https://chief-install.s3.eu-central-1.amazonaws.com/stack.yml.template | envsubst | tee /var/chief/stack.yml").Run()
				if err != nil {
					fmt.Println(errorStyle.Render("Error downloading stack.yml."))
					os.Exit(1)
				}

			}).
			Run()

		if err_download_stack != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_download_stack)))
			os.Exit(1)
		}
	} else {
		err := exec.Command("bash", "-c", "export DOMAIN="+domain+"; export PASSWORD="+password+"; export CHIEF_VERSION="+string(version)+"; curl -fsSL https://chief-install.s3.eu-central-1.amazonaws.com/stack.yml.template | envsubst | tee /var/chief/stack.yml").Run()

		if err != nil {
			fmt.Println(errorStyle.Render("Error downloading stack.yml."))
			os.Exit(1)
		}
	}

	if !cmd.Flags().Changed("cron") {
		err_download_caddyfile := spinner.New().
			Title("Downloading Caddyfile...").
			Action(func() {
				err := exec.Command("bash", "-c", "export DOMAIN="+domain+"; curl -fsSL https://chief-install.s3.eu-central-1.amazonaws.com/Caddyfile.template | envsubst | tee /var/chief/Caddyfile").Run()
				if err != nil {
					fmt.Println(errorStyle.Render("Error downloading Caddyfile."))
					os.Exit(1)
				}

			}).
			Run()

		if err_download_caddyfile != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_download_caddyfile)))
			os.Exit(1)
		}
	} else {
		err := exec.Command("bash", "-c", "export DOMAIN="+domain+"; curl -fsSL https://chief-install.s3.eu-central-1.amazonaws.com/Caddyfile.template | envsubst | tee /var/chief/Caddyfile").Run()
		if err != nil {
			fmt.Println(errorStyle.Render("Error downloading Caddyfile."))
			os.Exit(1)
		}
	}

	if !cmd.Flags().Changed("cron") {
		err_deploy_stack := spinner.New().
			Title("Deploying the Chief stack...").
			Action(func() {
				err := exec.Command("bash", "-c", "cd /var/chief && docker pull ghcr.io/chiefdeploy/controller:latest && docker stack deploy -c stack.yml --detach=true --resolve-image changed chief").Run()
				if err != nil {
					fmt.Println(errorStyle.Render("Error deploying stack."))
					os.Exit(1)
				}
			}).
			Run()

		if err_deploy_stack != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_deploy_stack)))
			os.Exit(1)
		}
	} else {
		err := exec.Command("bash", "-c", "cd /var/chief && docker pull ghcr.io/chiefdeploy/controller:latest && docker stack deploy -c stack.yml --detach=true --resolve-image changed chief").Run()
		if err != nil {
			fmt.Println(errorStyle.Render("Error deploying stack."))
			os.Exit(1)
		}
	}

	if !cmd.Flags().Changed("cron") {
		fmt.Println(installStyle.Render("Chief controller has been updated."))
	}
}
