package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/chiefdeploy/cli/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
)

var (
	installStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs the Chief controller.",
	Long:  "Installs the Chief controller.",
	Run: func(cmd *cobra.Command, args []string) {
		install()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func isDomainValid(domain string) error {
	if len(domain) < 3 {
		return errors.New(errorStyle.Render("Please use a valid domain."))
	}

	if !strings.Contains(domain, ".") {
		return errors.New(errorStyle.Render("Please use a valid domain."))
	}

	return nil
}

func createFolders() {
	err := os.MkdirAll("/var/chief", 0755)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func saveDomain(domain string) {
	viper.Set("domain", domain)
	err := viper.WriteConfig()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func saveAutomaticUpdates(automaticUpdates bool) {
	viper.Set("automatic_updates", automaticUpdates)
	err := viper.WriteConfig()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func savePassword(password string) {
	viper.Set("password", password)
	err := viper.WriteConfig()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func generatePassword() string {
	rand.Seed(uint64(time.Now().UnixNano()))
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	length := 48
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func setVersion(version string) {
	viper.Set("current_version", version)
	err := viper.WriteConfig()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}

func install() {
	var domain string
	var automaticUpdates bool = true
	var password string

	if os.Geteuid() != 0 {
		fmt.Println(errorStyle.Render("Please run this command with sudo. `sudo chief install`"))
		os.Exit(1)
	}

	if viper.GetString("domain") != "" {
		domain = viper.GetString("domain")
	}

	fmt.Println("Installing Chief controller...")

	lib.PrintLogo()

	err_check := exec.Command("docker", "stack", "ps", "chief").Run()

	if err_check == nil {
		fmt.Println(errorStyle.Render("Error: Chief controller is already installed. Please run `chief update` to update the controller."))
		os.Exit(1)
	}

	err_docker_check := spinner.New().
		Title("Checking Docker...").
		Action(func() {
			// Check Docker Version
			dockerVersion, err := exec.Command("docker", "--version").Output()
			if err != nil {
				fmt.Println(errorStyle.Render("Docker is not installed. Please install Docker and try again."))
				os.Exit(1)
			}
			fmt.Println(installStyle.Render(string(dockerVersion)))

			// Check Docker Compose Version
			dockerComposeVersion, err := exec.Command("docker", "compose", "version").Output()
			if err != nil {
				fmt.Println(errorStyle.Render("Docker Compose is not installed. Please install Docker Compose and try again."))
				os.Exit(1)
			}
			fmt.Println(installStyle.Render(string(dockerComposeVersion)))
		}).
		Run()

	if err_docker_check != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_docker_check)))
		os.Exit(1)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What domain would you like to assign to your controller? (Please use a valid domain and setup the DNS A record for it to point at this server!)").
				Prompt("Domain: ").
				Placeholder("hosting.yourdomain.com").
				Validate(isDomainValid).
				Value(&domain),
			huh.NewConfirm().
				Title("Would you like to enable automatic updates?").
				Description("If you enable automatic updates, Chief Controller will automatically update itself at 2am every day.").
				Affirmative("Yes!").
				Negative("No.").
				Value(&automaticUpdates),
		),
	)

	err := form.Run()

	if err != nil {
		log.Fatal(err)
	}

	if domain == "" {
		fmt.Println(errorStyle.Render("Error: Domain cannot be empty."))
		os.Exit(1)
	}

	if !strings.Contains(domain, ".") || len(domain) < 3 {
		fmt.Println(errorStyle.Render("Error: Please use a valid domain."))
		os.Exit(1)
	}

	fmt.Println("Domain:", domain)

	createFolders()

	err_save_domain := spinner.New().
		Title("Saving domain...").
		Action(func() {
			saveDomain(domain)
		}).
		Run()

	if err_save_domain != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_save_domain)))
		os.Exit(1)
	}

	err_save_automatic_updates := spinner.New().
		Title("Saving automatic updates...").
		Action(func() {
			saveAutomaticUpdates(automaticUpdates)
		}).
		Run()

	if err_save_automatic_updates != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_save_automatic_updates)))
		os.Exit(1)
	}

	var automaticUpdatesString string

	if automaticUpdates {
		automaticUpdatesString = "Yes"

		err := exec.Command("bash", "-c", "echo \"0 3 * * * root /usr/local/bin/chief update\" > /etc/cron.d/chief_update").Run()
		if err != nil {
			fmt.Println(errorStyle.Render("Error setting up automatic updates. Please manually run 'chief update' as root to update Chief Controller."))
			os.Exit(1)
		}
	} else {
		automaticUpdatesString = "No"

		err := exec.Command("bash", "-c", "rm -f /etc/cron.d/chief_update").Run()
		if err != nil {
			fmt.Println(errorStyle.Render("Error removing automatic updates. Please manually run 'rm -f /etc/cron.d/chief_update' as root to remove Chief Controller's automatic updates."))
			os.Exit(1)
		}
	}

	fmt.Println("\n\r")

	fmt.Printf("Automatic updates: %s\n", automaticUpdatesString)

	fmt.Println("\n\r")

	err_save_password := spinner.New().
		Title("Generating db password...").
		Action(func() {
			password = generatePassword()
			savePassword(password)
		}).
		Run()

	if err_save_password != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_save_password)))
		os.Exit(1)
	}

	if password == "" {
		fmt.Println(errorStyle.Render("Error: Generating password failed."))
		os.Exit(1)
	}

	version, err := exec.Command("bash", "-c", "curl -fsSL https://install.chiefdeploy.com/version").Output()

	if err != nil {
		fmt.Println(errorStyle.Render("Error: Unable to retrieve Chief version."))
		os.Exit(1)
	}

	fmt.Println(installStyle.Render("Chief version: " + string(version)))

	setVersion(string(strings.ReplaceAll(string(version), "\n", "")))

	fmt.Println("")

	err_swarm_init := spinner.New().
		Title("Initializing Docker Swarm...").
		Action(func() {
			err := exec.Command("bash", "-c", "docker swarm init --advertise-addr 127.0.0.1").Run()
			if err != nil {
				fmt.Println(errorStyle.Render("Error initializing Docker Swarm."))
				os.Exit(1)
			}
		}).
		Run()

	if err_swarm_init != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err_swarm_init)))
		os.Exit(1)
	}

	err_download_stack := spinner.New().
		Title("Downloading stack.yml...").
		Action(func() {
			// run bash command to download stack.yml and replace DOMAIN and PASSWORD
			err := exec.Command("bash", "-c", "export DOMAIN="+domain+"; export PASSWORD="+password+"; export CHIEF_VERSION="+string(strings.ReplaceAll(string(version), "\n", ""))+"; curl -fsSL https://chief-install.s3.eu-central-1.amazonaws.com/stack.yml.template | envsubst | tee /var/chief/stack.yml").Run()
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

	err_deploy_stack := spinner.New().
		Title("Deploying the Chief stack... (this may take a few minutes)").
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

	fmt.Println(installStyle.Render("Chief controller is now running. To access it, visit https://" + domain))
}
