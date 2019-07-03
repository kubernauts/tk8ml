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

	"github.com/kubernauts/tk8ml/pkg/common"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var kubeflow, k8s, chainerOperator, katib bool

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
			kubeConfig := common.GetKubeConfig()
			common.CheckKubectl(kubeConfig)
			installChainerOperator()
			os.Exit(0)
		}

		if katib {
			kubeConfig := common.GetKubeConfig()
			common.CheckKubectl(kubeConfig)
			installKatib()
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

func installChainerOperator() {
	fmt.Println(aurora.Cyan("Enter the KSONNET_APP directory path"))
	var ksAppDir, ksEnvVar string
	fmt.Scanln(&ksAppDir)
	fmt.Println(aurora.Cyan("Enter the value for ENVIRONMENT variable. By default, it is set to \"default\"."))
	fmt.Scanln(&ksEnvVar)
	os.Setenv("ENVIRONMENT", ksEnvVar)
	fmt.Println("Installing Chainer Operator package")
	ksInstallPkg := exec.Command("ks", "pkg", "install", "kubeflow/chainer-job")
	ksInstallPkg.Dir = ksAppDir
	stdout, err := ksInstallPkg.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksInstallPkg.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := ksInstallPkg.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	fmt.Println("Generating chainer-operator by ksonnet")
	ksGen := exec.Command("ks", "generate", "chainer-operator", "chainer-operator")
	ksGen.Dir = ksAppDir
	stdout, err = ksGen.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksGen.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := ksGen.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	fmt.Println("Applying chainer-operator config")
	ksApply := exec.Command("ks", "apply", os.ExpandEnv("$ENVIRONMENT"), "-c", "chainer-operator")
	ksApply.Dir = ksAppDir
	stdout, err = ksApply.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksApply.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := ksApply.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	fmt.Println(aurora.Green("Chainer Operator has been deployed successfully"))
}

func installKatib() {
	fmt.Println(aurora.Cyan("Enter the KSONNET_APP directory path"))
	var ksAppDir string
	fmt.Scanln(&ksAppDir)

	fmt.Println("Setting KF_ENV env var to default")
	os.Setenv("$KF_ENV", "default")
	kfEnvVar := "default"
	ksEnvSet := exec.Command("ks", "env", "set", os.ExpandEnv("$KF_ENV"), "--namespace=kubeflow")

	ksEnvSet.Dir = ksAppDir
	stdout, err := ksEnvSet.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksEnvSet.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := ksEnvSet.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	fmt.Println("Adding kubeflow registry")
	kubeflowUrl := "github.com/kubeflow/kubeflow/tree/master/kubeflow"
	ksAddRegistry := exec.Command("ks", "registry", "add", "kubeflow", kubeflowUrl)
	ksAddRegistry.Dir = ksAppDir
	stdout, err = ksAddRegistry.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksAddRegistry.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := ksAddRegistry.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	// Install TF Job Operator
	installTfJob(ksAppDir, kfEnvVar)

	// Install Pytorch
	installPytorch(ksAppDir, kfEnvVar)

	fmt.Println("Installing Katib")
	pkgInstallKatib := exec.Command("ks", "pkg", "install", "kubeflow/katib")
	pkgInstallKatib.Dir = ksAppDir
	stdout, err = pkgInstallKatib.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := pkgInstallKatib.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := pkgInstallKatib.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	generateKatib := exec.Command("ks", "generate", "katib", "katib")
	generateKatib.Dir = ksAppDir
	stdout, err = generateKatib.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := generateKatib.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := generateKatib.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	applykatib := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "katib")
	generateKatib.Dir = ksAppDir
	stdout, err = applykatib.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := applykatib.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := applykatib.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
}

func installTfJob(ksAppDir string, kfEnvVar string) {
	fmt.Println("Installing Tensorflow job operator")
	tfTrainingInstall := exec.Command("ks", "pkg", "install", "kubeflow/tf-training")
	tfTrainingInstall.Dir = ksAppDir
	stdout, err := tfTrainingInstall.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := tfTrainingInstall.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := tfTrainingInstall.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	kubeflowCommonInstall := exec.Command("ks", "pkg", "install", "kubeflow/common")
	kubeflowCommonInstall.Dir = ksAppDir
	stdout, err = kubeflowCommonInstall.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := kubeflowCommonInstall.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := kubeflowCommonInstall.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	generateTfOperator := exec.Command("ks", "generate", "tf-job-operator", "tf-job-operator")
	generateTfOperator.Dir = ksAppDir
	stdout, err = generateTfOperator.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := generateTfOperator.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := generateTfOperator.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	os.Setenv("$KF_ENV", kfEnvVar)
	ksApplyTfJob := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "tf-job-operator")
	ksApplyTfJob.Dir = ksAppDir
	stdout, err = ksApplyTfJob.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := ksApplyTfJob.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := ksApplyTfJob.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

}

func installPytorch(ksAppDir string, kfEnvVar string) {
	fmt.Println("Installing PyTorch Job operator")
	pkgInstallPytorch := exec.Command("ks", "pkg", "install", "kubeflow/pytorch-job")
	pkgInstallPytorch.Dir = ksAppDir
	stdout, err := pkgInstallPytorch.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := pkgInstallPytorch.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := pkgInstallPytorch.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}

	generatePytorch := exec.Command("ks", "generate", "pytorch-operator", "pytorch-operator")
	generatePytorch.Dir = ksAppDir
	stdout, err = generatePytorch.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := generatePytorch.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := generatePytorch.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	os.Setenv("$KF_ENV", kfEnvVar)
	applyPytorch := exec.Command("ks", "apply", os.ExpandEnv("$KF_ENV"), "-c", "pytorch-operator")
	applyPytorch.Dir = ksAppDir
	stdout, err = applyPytorch.StdoutPipe()
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	scanner = bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := applyPytorch.Start(); err != nil {
		log.Fatal(aurora.Red(err))
	}
	if err := applyPytorch.Wait(); err != nil {
		log.Fatal(aurora.Red(err))
	}
}
