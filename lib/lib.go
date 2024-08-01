package lib

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/fynelabs/selfupdate"
)

var (
	logoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0f95d4")).PaddingTop(1).PaddingBottom(1)
)

var logo = ` ██████╗██╗  ██╗██╗███████╗███████╗
██╔════╝██║  ██║██║██╔════╝██╔════╝
██║     ███████║██║█████╗  █████╗  
██║     ██╔══██║██║██╔══╝  ██╔══╝  
╚██████╗██║  ██║██║███████╗██║     
 ╚═════╝╚═╝  ╚═╝╚═╝╚══════╝╚═╝`

func PrintLogo() {
	fmt.Println(logoStyle.Render(logo))
}

func GetExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath + "/chief"), nil
}

func GetChecksumOfFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GetCLIVersion() string {
	exec_path, err := GetExecutablePath()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	checksum, err := GetChecksumOfFile(exec_path)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	return checksum
}

func UpdateCLI(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		// error handling
		fmt.Println("Error updating Chief CLI.")
		os.Exit(1)
	}

	fmt.Println("Chief CLI has been updated.")

	return err
}
