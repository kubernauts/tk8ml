/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var kubeflow, k8s bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kubeflow",
	Long:  `This command will setup Kubeflow on the kubernetes cluster`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

var kubeFlowCmd = &cobra.Command{
	Use:   "kubeflow",
	Short: "Install Kubeflow",
	Long:  `This command will setup Kubeflow on the kubernetes cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Inside kubeflow command")
		fmt.Println("install called")
		checkKfctl()
		kubeConfig := getKubeConfig()
		checkKubectl(kubeConfig)
		kubeFlowInstall(kubeConfig)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(kubeFlowCmd)
	kubeFlowCmd.Flags().BoolVarP(&k8s, "k8s", "", false, "Deploy Kubeflow on an existing Kubernetes cluster")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkKfctl() {
	fmt.Println("Checking if kfctl command exists on the system and is added in $PATH")
	kfctlDir, err := exec.LookPath("kfctl")
	if err != nil {
		log.Fatal("kfctl command not found. Please check if kfctl is installed correctly.")
	}
	fmt.Printf("Found kfctl at %s\n", kfctlDir)

}

func getKubeConfig() string {
	// Get kubeconfig
	fmt.Println("Please enter the path to your kubeconfig:")
	var kubeConfig string
	fmt.Scanln(&kubeConfig)
	fmt.Printf("path: %s\n", kubeConfig)
	if _, err := os.Stat(kubeConfig); err != nil {
		fmt.Println("Kubeconfig file not found, kindly check")
		os.Exit(1)
	}
	return kubeConfig
}

func checkKubectl(kubeConfig string) {
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

func kubeFlowInstall(kubeConfig string) {
	fmt.Println("Setting KUBECONFIG environment variable.")
	err := os.Setenv("KUBECONFIG", kubeConfig)
	if err != nil {
		log.Fatal("Unable to set KUBECONFIG env var")
	}
	fmt.Println("Please enter the directory where you want to setup Kubeflow.")
	var kfDir string
	fmt.Scanln(&kfDir)
	fmt.Printf("Kubeflow install path: %s", kfDir)
	fmt.Println("Kubeflow directory exists on the system. Proceeding with the installation.")
	err = os.Setenv("KFAPP", kfDir)

	if err != nil {
		log.Fatal("Unable to set env var KFAPP.")
	}

	fmt.Println("Creating KFAPP directory if it doesn't exist")
	if _, err := os.Stat(kfDir); os.IsNotExist(err) {
		os.Mkdir(kfDir, 0755)
	}
	if err != nil {
		fmt.Errorf("Cannot create the directory: %s", kfDir)
	}
	fmt.Println("kfctl","init","${KFAPP}")
	fmt.Println("env get", os.Getenv("KFAPP"))

	initCmd := exec.Command("kfctl","init","${KFAPP}")
	initCmd.Env = os.Environ()

	if err != nil {
		log.Fatal("Cannot initialise with kfctl. Exiting.")
	}

	fmt.Println("Starting kfctl generate")
	kfGenerateCmd := exec.Command("kfctl", "generate", "all", "-V")
	kfGenerateCmd.Dir = kfDir
	stdout, err := kfGenerateCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfGenerateCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := kfGenerateCmd.Wait(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting kfctl apply")
	kfApplyCmd := exec.Command("kfctl", "apply", "all", "-V")
	kfApplyCmd.Dir = kfDir
	stdout, err = kfApplyCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfApplyCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := kfApplyCmd.Wait(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Checking if all the resources are deployed in the namespace kubeflow")
	_, err = exec.Command("kubectl", "-n", "kubeflow", "get", "all").Output()
	if err != nil {
		log.Fatal("Kubeflow is not deployed successfully.")
	}
	fmt.Println("Successfully deployed Kubeflow. Have a pleasant time creating ML workflows.")
}
