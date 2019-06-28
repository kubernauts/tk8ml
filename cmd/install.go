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
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/go-github/github"
	"github.com/kubernauts/tk8ml/pkg/common"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var kubeflow, k8s, minikube bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Kubeflow",
	Long:  `This command will setup Kubeflow on the kubernetes cluster`,
	Args:  cobra.ExactArgs(1),
	Run:   func(cmd *cobra.Command, args []string) {},
}

var kubeFlowCmd = &cobra.Command{
	Use:   "kubeflow",
	Short: "Install Kubeflow",
	Long:  `This command will setup Kubeflow on the kubernetes cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if k8s {
			common.CheckKfctl()
			kubeConfig := common.GetKubeConfig()
			common.CheckKubectl(kubeConfig)
			kubeFlowInstall(kubeConfig)
			os.Exit(0)
		}

		if minikube {
			kubeConfig := common.GetKubeConfig()
			common.CheckKubectl(kubeConfig)
			kubeFlowInstallMinikube()
		}

		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(kubeFlowCmd)
	kubeFlowCmd.Flags().BoolVarP(&k8s, "k8s", "", false, "Deploy Kubeflow on an existing Kubernetes cluster")
	kubeFlowCmd.Flags().BoolVarP(&minikube, "minikube", "", false, "Deploy Kubeflow on existing minikube installation")
}

func kubeFlowInstall(kubeConfig string) {
	fmt.Println("Setting KUBECONFIG environment variable.")
	err := os.Setenv("KUBECONFIG", kubeConfig)
	if err != nil {
		log.Fatal(aurora.Red("Unable to set KUBECONFIG env var"))
	}
	fmt.Println(aurora.Cyan("Please enter the directory where you want to setup Kubeflow."))
	var kfDir string
	fmt.Scanln(&kfDir)
	fmt.Printf("Kubeflow install path: %s\n", kfDir)
	err = os.Setenv("KFAPP", kfDir)

	if err != nil {
		log.Fatal(aurora.Red("Unable to set env var KFAPP."))
	}

	_, err = exec.Command("kfctl", "init", os.ExpandEnv("$KFAPP")).Output()

	if err != nil {
		log.Fatal(aurora.Red("Cannot initialise with kfctl. Exiting."))
	}

	fmt.Println("Starting kfctl generate")
	kfGenerateCmd := exec.Command("kfctl", "generate", "all", "-V")
	kfGenerateCmd.Dir = kfDir
	stdout, err := kfGenerateCmd.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfGenerateCmd.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := kfGenerateCmd.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	fmt.Println("Starting kfctl apply")
	kfApplyCmd := exec.Command("kfctl", "apply", "all", "-V")
	kfApplyCmd.Dir = kfDir
	stdout, err = kfApplyCmd.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfApplyCmd.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := kfApplyCmd.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	fmt.Println("Checking if all the resources are deployed in the namespace kubeflow")
	verifyKubeflow := exec.Command("kubectl", "-n", "kubeflow", "get", "all")
	verifyKubeflow.Dir = kfDir
	stdout, err = verifyKubeflow.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := verifyKubeflow.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := verifyKubeflow.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	fmt.Println(aurora.Green("Successfully deployed Kubeflow. Have a pleasant time creating ML workflows."))
}

func kubeFlowInstallMinikube() {
	fmt.Println(aurora.Cyan("Please enter the directory where you want to download the Kubeflow setup script."))
	var kfScriptDir string
	_, err := fmt.Scanln(&kfScriptDir)
	if err != nil {
		log.Fatal(aurora.Red("Unable to read the input"), aurora.Red(err))
	}
	err = os.Setenv("KUBEFLOW_SRC", kfScriptDir)
	if err != nil {
		log.Fatal(aurora.Red("Unable to set environment variable KUBEFLOW_SRC"), aurora.Red(err))
	}
	fmt.Println("Fetching the latest Kubeflow tag from releases")
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.GetLatestRelease(context.Background(), "kubeflow", "kubeflow")
	if err != nil {
		fmt.Println("release error", err)
	}

	fetchKubeflowTag := *releases.TagName
	fmt.Println(fetchKubeflowTag)
	err = os.Setenv("KUBEFLOW_TAG", fetchKubeflowTag)
	if err != nil {
		log.Fatal(aurora.Red("Unable to set environment variable KUBEFLOW_TAG"), aurora.Red(err))
	}
	downloadUrl := "https://raw.githubusercontent.com/kubeflow/kubeflow/" + os.ExpandEnv("$KUBEFLOW_TAG") + "/scripts/download.sh"
	_, err = exec.Command("wget", "-P", kfScriptDir, downloadUrl).Output()
	fmt.Println("/bin/sh", "download.sh")
	if err != nil {
		log.Fatal(aurora.Red("Unable to download the kubeflow download script"), os.Stderr)
	}
	kfScriptDownloadCmd := exec.Command("/bin/sh", "download.sh")

	kfScriptDownloadCmd.Dir = kfScriptDir

	_, err = kfScriptDownloadCmd.Output()
	if err != nil {
		fmt.Println("test")
	}

	fmt.Println(aurora.Cyan("Please enter the directory where you want to setup Kubeflow."))
	var kfDir string
	fmt.Scanln(&kfDir)
	fmt.Printf("Kubeflow install path: %s\n", kfDir)
	err = os.Setenv("KFAPP", kfDir)

	if err != nil {
		log.Fatal(aurora.Red("Unable to set env var KFAPP."))
	}

	os.Setenv("$KUBEFLOW_REPO", "$KUBEFLOW_SRC")
	kfInitCmd := exec.Command("kfctl", "init", os.ExpandEnv("$KFAPP"), "--platform", "minikube")

	stdout, err := kfInitCmd.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfInitCmd.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := kfInitCmd.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	fmt.Println("Starting kfctl generate")
	kfGenerateCmd := exec.Command("kfctl", "generate", "all", "-V")
	kfGenerateCmd.Dir = kfDir
	stdout, err = kfGenerateCmd.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfGenerateCmd.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := kfGenerateCmd.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	fmt.Println("Starting kfctl apply")
	kfApplyCmd := exec.Command("kfctl", "apply", "all", "-V")
	kfApplyCmd.Dir = kfDir
	stdout, err = kfApplyCmd.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfApplyCmd.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := kfApplyCmd.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	fmt.Println(aurora.Green("Successfully deployed Kubeflow. Have a pleasant time creating ML workflows."))

}
