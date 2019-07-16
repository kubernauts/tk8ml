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
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/kubernauts/tk8ml/pkg/common"
	. "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var kubeflow, k8s, chainerOperator, katib, modeldb bool

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
		common.CheckKfctl()
		kubeConfig := common.GetKubeConfig()
		common.CheckKubectl(kubeConfig)
		kubeFlowInstall(kubeConfig)
		os.Exit(0)
	},
}

var kubeFlowComponentCmd = &cobra.Command{
	Use:   "kubeflow-component",
	Short: "Installs Kubeflow components",
	Long:  `This command will install different kubeflow components`,
	Run: func(cmd *cobra.Command, args []string) {
		if chainerOperator {
			common.CheckKfctl()
			installChainerOperator()
			os.Exit(0)
		}

		if katib {
			installKatib()
			os.Exit(0)

		}

		if modeldb {
			installModelDb()
			os.Exit(0)
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
	installCmd.AddCommand(kubeFlowComponentCmd)
	kubeFlowCmd.Flags().BoolVarP(&k8s, "k8s", "", false, "Deploy Kubeflow on an existing Kubernetes cluster")
	kubeFlowComponentCmd.Flags().BoolVarP(&chainerOperator, "chainer-operator", "", false, "Deploy Chainer Operator")
	kubeFlowComponentCmd.Flags().BoolVarP(&katib, "katib", "", false, "Deploy Katib")
	kubeFlowComponentCmd.Flags().BoolVarP(&modeldb, "modeldb", "", false, "Deploy ModelDB")


}

func kubeFlowInstall(kubeConfig string) {
	fmt.Println("Setting KUBECONFIG environment variable.")
	err := os.Setenv("KUBECONFIG", kubeConfig)
	if err != nil {
		log.Fatal(Red("Unable to set KUBECONFIG env var"))
	}
	fmt.Println(Cyan("Please enter the directory where you want to setup Kubeflow."))
	var kfDir string
	fmt.Scanln(&kfDir)
	fmt.Printf("Kubeflow install path: %s\n", kfDir)
	err = os.Setenv("KFAPP", kfDir)

	if err != nil {
		log.Fatal(Red("Unable to set env var KFAPP."))
	}

	_, err = exec.Command("kfctl", "init", os.ExpandEnv("$KFAPP")).Output()

	if err != nil {
		log.Fatal(Red("Cannot initialise with kfctl. Exiting."))
	}

	fmt.Println("Starting kfctl generate")
	kfGenerateCmd := exec.Command("kfctl", "generate", "all", "-V")
	kfGenerateCmd.Dir = kfDir
	stdout, err := kfGenerateCmd.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfGenerateCmd.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := kfGenerateCmd.Wait(); err != nil {
		log.Fatal(Red(err))
	}

	fmt.Println("Starting kfctl apply")
	kfApplyCmd := exec.Command("kfctl", "apply", "all", "-V")
	kfApplyCmd.Dir = kfDir
	stdout, err = kfApplyCmd.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kfApplyCmd.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := kfApplyCmd.Wait(); err != nil {
		log.Fatal(Red(err))
	}

	fmt.Println("Checking if all the resources are deployed in the namespace kubeflow")
	verifyKubeflow := exec.Command("kubectl", "-n", "kubeflow", "get", "all")
	verifyKubeflow.Dir = kfDir
	stdout, err = verifyKubeflow.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := verifyKubeflow.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := verifyKubeflow.Wait(); err != nil {
		log.Fatal(Red(err))
	}
	fmt.Println(Green("Successfully deployed Kubeflow. Have a pleasant time creating ML workflows."))
}

func installChainerOperator() {
	fmt.Println(Cyan("Enter the KSONNET_APP directory path"))
	var ksAppDir, ksEnvVar string
	fmt.Scanln(&ksAppDir)

	componentName := "chainer-job"
	common.CheckComponentExist(componentName, ksAppDir)
	fmt.Println(Cyan("\nEnter the value for ENVIRONMENT variable. By default, it is set to \"default\"."))
	fmt.Scanln(&ksEnvVar)
	os.Setenv("ENVIRONMENT", ksEnvVar)
	fmt.Println("Installing Chainer Operator package")

	pkgName := "kubeflow/chainer-job"

	// install pkg
	common.KsPkgInstall(pkgName, ksAppDir)

	componentGenerateName := "chainer-operator"

	// generate component with ksonnet
	common.ComponentGenerate(componentGenerateName, ksAppDir)

	fmt.Println("Applying chainer-operator config")
	ksApply := exec.Command("ks", "apply", os.ExpandEnv("$ENVIRONMENT"), "-c", "chainer-operator")
	ksApply.Dir = ksAppDir
	stdout, err := ksApply.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksApply.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := ksApply.Wait(); err != nil {
		log.Fatal(Red(err))
	}
	fmt.Println(Green("Chainer Operator has been deployed successfully"))
}

func installKatib() {
	var outb, errb bytes.Buffer
	fmt.Println(Cyan("Enter the KSONNET_APP directory path"))
	var ksAppDir string
	fmt.Scanln(&ksAppDir)
	componentName := "katib"
	common.CheckComponentExist(componentName, ksAppDir)

	fmt.Println("Setting KF_ENV env var to default")
	os.Setenv("$KF_ENV", "default")
	kfEnvVar := "default"
	ksEnvSet := exec.Command("ks", "env", "set", os.ExpandEnv("$KF_ENV"), "--namespace=kubeflow")

	ksEnvSet.Dir = ksAppDir
	stdout, err := ksEnvSet.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksEnvSet.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := ksEnvSet.Wait(); err != nil {
		log.Fatal(Red(err))
	}

	fmt.Println("Adding kubeflow registry")
	kubeflowUrl := "github.com/kubeflow/kubeflow/tree/master/kubeflow"
	ksAddRegistry := exec.Command("ks", "registry", "add", "kubeflow", kubeflowUrl)
	ksAddRegistry.Dir = ksAppDir
	ksAddRegistry.Stdout = &outb
	ksAddRegistry.Stderr = &errb
	err = ksAddRegistry.Run()
	if err != nil {
		log.Fatal(Red(errb.String()))
	}

	// Install TF Job Operator
	installTfJob(ksAppDir, kfEnvVar)

	// Install Pytorch
	installPytorch(ksAppDir, kfEnvVar)

	fmt.Println("Installing Katib")
	pkgName := "kubeflow/katib"

	// install pkg
	common.KsPkgInstall(pkgName, ksAppDir)

	componentGenerateName := "katib"

	// generate component with ksonnet
	common.ComponentGenerate(componentGenerateName, ksAppDir)

	applykatib := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "katib")
	applykatib.Dir = ksAppDir
	stdout, err = applykatib.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := applykatib.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := applykatib.Wait(); err != nil {
		log.Fatal(Red(err))
	}
	fmt.Println(Green("Katib has been deployed successfully"))

}

func installTfJob(ksAppDir string, kfEnvVar string) {
	componentName := "tf-training"
	common.CheckComponentExist(componentName, ksAppDir)

	fmt.Printf("App dir %s", ksAppDir)
	fmt.Println("ks", "pkg", "install", "kubeflow/tf-training")

	pkgName := "kubeflow/tf-training"

	// install pkg
	common.KsPkgInstall(pkgName, ksAppDir)

	fmt.Println("Installing kubeflow/common package")
	pkgName = "kubeflow/common"

	// install pkg
	common.KsPkgInstall(pkgName, ksAppDir)

	componentGenerateName := "tf-job-operator"

	// generate component with ksonnet
	common.ComponentGenerate(componentGenerateName, ksAppDir)

	os.Setenv("$KF_ENV", kfEnvVar)
	fmt.Println("apply")
	ksApplyTfJob := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "tf-job-operator")
	ksApplyTfJob.Dir = ksAppDir
	stdout, err := ksApplyTfJob.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksApplyTfJob.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := ksApplyTfJob.Wait(); err != nil {
		log.Fatal(Red(err))
	}
	fmt.Println(Green("TFJob has been deployed successfully"))


}

func installPytorch(ksAppDir string, kfEnvVar string) {
	componentName := "pytorch-job"
	common.CheckComponentExist(componentName, ksAppDir)

	fmt.Println("Pytorch package is not installed. Installing.")
	fmt.Println("Installing PyTorch Job operator")

	pkgName := "kubeflow/pytorch-job"

	// install pkg
	common.KsPkgInstall(pkgName, ksAppDir)

	componentGenerateName := "pytorch-operator"

	// generate component with ksonnet
	common.ComponentGenerate(componentGenerateName, ksAppDir)

	os.Setenv("$KF_ENV", kfEnvVar)
	fmt.Println(os.ExpandEnv("$KF_ENV"))
	fmt.Println("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "pytorch-operator")
	applyPytorch := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "pytorch-operator")
	applyPytorch.Dir = ksAppDir
	stdout, err := applyPytorch.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := applyPytorch.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := applyPytorch.Wait(); err != nil {
		log.Fatal(Red(err))
	}
	fmt.Println(Green("PyTorch has been deployed successfully"))

}

func installModelDb() {
	fmt.Println(Cyan("Enter the KSONNET_APP directory path"))
	var ksAppDir string
	fmt.Scanln(&ksAppDir)
	componentName := "modeldb"
	common.CheckComponentExist(componentName, ksAppDir)

	componentGenerateName := "modeldb"

	// generate component with ksonnet
	common.ComponentGenerate(componentGenerateName, ksAppDir)

	fmt.Println("Setting KF_ENV env var to default")
	os.Setenv("$KF_ENV", "default")

	applyModelDb := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "modeldb")
	applyModelDb.Dir = ksAppDir
	stdout, err := applyModelDb.StdoutPipe()
	if err != nil {
		log.Fatal(Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := applyModelDb.Start(); err != nil {
		log.Fatal(Red(err))
	}
	if err := applyModelDb.Wait(); err != nil {
		log.Fatal(Red(err))
	}
	fmt.Println(Green("ModelDB has been deployed successfully"))
}