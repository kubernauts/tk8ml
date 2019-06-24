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
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete kubeflow installation",
	Long:  `This command will delete the kubeflow installation.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var DeletekubeFlowCmd = &cobra.Command{
	Use:   "kubeflow",
	Short: "Delete kubeflow installation",
	Long:  `This command will delete the kubeflow installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if all {
			common.CheckKfctl()
			kubeConfig := common.GetKubeConfig()
			common.CheckKubectl(kubeConfig)
			deleteFlag := "all"
			kubeFlowDelete(kubeConfig, deleteFlag)
			os.Exit(0)
		}

		if deleteStorage {
			common.CheckKfctl()
			kubeConfig := common.GetKubeConfig()
			common.CheckKubectl(kubeConfig)
			deleteFlag := "deleteStorage"
			kubeFlowDelete(kubeConfig, deleteFlag)
			os.Exit(0)
		}
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

var all bool
var deleteStorage bool

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(DeletekubeFlowCmd)
	DeletekubeFlowCmd.Flags().BoolVarP(&all, "all", "", false, "Preserve storage, delete everything")
	DeletekubeFlowCmd.Flags().BoolVarP(&deleteStorage, "delete_storage", "", false, "Delete everything along with storage")

}

func kubeFlowDelete(kubeConfig string, deleteFlag string) {
	fmt.Println("Please enter the directory where you had setup Kubeflow.")
	var kfDir string
	fmt.Scanln(&kfDir)
	fmt.Printf("Kubeflow install path: %s", kfDir)
	//fmt.Println("Kubeflow directory exists on the system. Proceeding with the installation.")
	err := os.Setenv("KFAPP", kfDir)

	if err != nil {
		log.Fatal("Unable to set env var KFAPP.")
	}

	fmt.Println("Deleting kubeflow installation")
	fmt.Println("delete flag", deleteFlag)
	if deleteFlag == "all" {
		fmt.Println("Deleting everything and preserving storage containing metadata from ML pipelines")
		kfDeleteCmd := exec.Command("kfctl", "delete", "all")
		kfDeleteCmd.Dir = kfDir
		stdout, err := kfDeleteCmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(stdout)
		go func() {
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}()
		if err := kfDeleteCmd.Start(); err != nil {
			log.Fatal(err)
		}
		if err := kfDeleteCmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}

	if deleteFlag == "deleteStorage" {
		fmt.Println("Deleting everything including persistent storage")
		kfDeleteCmd := exec.Command("kfctl", "delete", "all", "--delete_storage")
		kfDeleteCmd.Dir = kfDir
		stdout, err := kfDeleteCmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(stdout)
		go func() {
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}()
		if err := kfDeleteCmd.Start(); err != nil {
			log.Fatal(err)
		}
		if err := kfDeleteCmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}
}
