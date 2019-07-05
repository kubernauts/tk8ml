package common

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"

	. "github.com/logrusorgru/aurora"
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
	fmt.Println(Cyan("Please enter the path to your kubeconfig:"))
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

func CheckComponentExists(componentName string, ksAppDir string) {
	var outb, errb bytes.Buffer
	fmt.Printf("Checking if %s already exists.", componentName)
	checkComponentExists := exec.Command("ks", "pkg", "list")
	checkComponentExists.Dir = ksAppDir
	checkComponentExists.Stdout = &outb
	checkComponentExists.Stderr = &errb
	_ = checkComponentExists.Run()
	match, _ := regexp.MatchString(componentName+"\\s+\\*", outb.String())
	if match {
		Magenta(fmt.Sprintf("%s already exists. Exiting.", componentName))
		os.Exit(0)
	}
	Magenta(fmt.Sprintf("%s is not installed. Installing.", componentName))
}

func ComponentGenerate(componentGenerateName string, ksAppDir string) {
	var outb, errb bytes.Buffer
	fmt.Printf("Generating %s by ksonnet", componentGenerateName)
	ksGen := exec.Command("ks", "generate", componentGenerateName, componentGenerateName)
	ksGen.Dir = ksAppDir

	ksGen.Dir = ksAppDir
	ksGen.Stdout = &outb
	ksGen.Stderr = &errb
	err := ksGen.Run()
	if err != nil {
		log.Fatal(Red(errb.String()))
	}

}

func KsPkgInstall(pkgName string, ksAppDir string) {
	var outb, errb bytes.Buffer
	pkgInstall := exec.Command("ks", "pkg", "install", pkgName)
	pkgInstall.Dir = ksAppDir
	pkgInstall.Stdout = &outb
	pkgInstall.Stderr = &errb
	err := pkgInstall.Run()
	if err != nil {
		log.Fatal(Red(errb.String()))
	}

}