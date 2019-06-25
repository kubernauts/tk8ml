package common

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"log"
	"os"
	"os/exec"
)

var (
	Name string
	// GITCOMMIT will hold the commit SHA to be used in the version command.
	GITCOMMIT = "0"
	// VERSION will hold the version number to be used in the version command.
	VERSION = "dev"
)

func GetKubeConfig() string {
	// Get kubeconfig
	fmt.Println(aurora.Cyan("Please enter the path to your kubeconfig:"))
	var kubeConfig string
	fmt.Scanln(&kubeConfig)
	fmt.Printf("path: %s\n", kubeConfig)
	if _, err := os.Stat(kubeConfig); err != nil {
		fmt.Println("Kubeconfig file not found, kindly check")
		os.Exit(1)
	}
	return kubeConfig
}

func CheckKubectl(kubeConfig string) {
	/*This function is used to check the whether kubectl command is installed &
	  it works with the kubeConfig provided
	*/
	kctlDir, err := exec.LookPath("kubectl")
	if err != nil {
		log.Fatal("kubectl command not found. Please check if kubectl is installed")
	}
	fmt.Printf("Found kubectl at %s\n", kctlDir)
	kver, err := exec.Command("kubectl", "--kubeconfig", kubeConfig, "version", "--short").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(string(kver))
}

func CheckKfctl() {
	fmt.Println("Checking if kfctl command exists on the system and is added in $PATH")
	kfctlDir, err := exec.LookPath("kfctl")
	if err != nil {
		log.Fatal("kfctl command not found. Please check if kfctl is installed correctly.")
	}
	fmt.Printf("Found kfctl at %s\n", kfctlDir)

}
