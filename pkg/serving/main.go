package serving

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kyokomi/emoji"
	. "github.com/logrusorgru/aurora"
	"github.com/spf13/viper"
)

type TfServingConfig struct {
	InstallIstio     bool
	DeploymentName   string
	ServiceName      string
	ModelName        string
	ServiceType      string
	VersionName      string
	ModelBasePath    string
	GcpSecretName    string
	AwsSecretname    string
	InjectIstio      bool
	S3Enable         bool
	S3SecretName     string
	S3AwsRegion      string
	S3UseHttps       bool
	S3VerifySslCerts bool
	S3EndpointUrl    string
	NumGpus          int
	ModelLocation    string
}

func GetTfServingConfig() TfServingConfig {
	ReadViperConfigFile("config")
	return TfServingConfig{
		InstallIstio:     viper.GetBool("kf-components.serving.tf-serving.install-istio"),
		DeploymentName:   viper.GetString("kf-components.serving.tf-serving.tf-serving-deployment-name"),
		ServiceName:      viper.GetString("kf-components.serving.tf-serving.tf-serving-service-name"),
		ModelName:        viper.GetString("kf-components.serving.tf-serving.model-name"),
		ServiceType:      viper.GetString("kf-components.serving.tf-serving.service-type"),
		VersionName:      viper.GetString("kf-components.serving.tf-serving.version-name"),
		ModelBasePath:    viper.GetString("kf-components.serving.tf-serving.model-base-path"),
		GcpSecretName:    viper.GetString("kf-components.serving.tf-serving.gcp-secret-name"),
		AwsSecretname:    viper.GetString("kf-components.serving.tf-serving.aws=secret-name"),
		InjectIstio:      viper.GetBool("kf-components.serving.tf-serving.inject-istio"),
		S3Enable:         viper.GetBool("kf-components.serving.tf-serving.s3-enable"),
		S3SecretName:     viper.GetString("kf-components.serving.tf-serving.s3-secret-name"),
		S3AwsRegion:      viper.GetString("kf-components.serving.tf-serving.s3-aws-region"),
		S3UseHttps:       viper.GetBool("kf-components.serving.tf-serving.s3-use-https"),
		S3VerifySslCerts: viper.GetBool("kf-components.serving.tf-serving.s3-verify-ssl-certs"),
		S3EndpointUrl:    viper.GetString("kf-components.serving.tf-serving.s3-endpoint-url"),
		NumGpus:          viper.GetInt("kf-components.serving.tf-serving.num-gpus"),
		ModelLocation:    viper.GetString("kf-components.serving.tf-serving.model-location"),
	}
}

// ReadViperConfigFile is define the config paths and read the configuration file.
func ReadViperConfigFile(configName string) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/tk8ml")
	verr := viper.ReadInConfig() // Find and read the config file.
	if verr != nil {             // Handle errors reading the config file.
		log.Fatalln(verr)
	}
}

func ConfigureTfServing() {
	fmt.Println(Cyan("Enter the KSONNET_APP directory path"))
	var ksAppDir string
	fmt.Scanln(&ksAppDir)
	fmt.Println("Inside ConfigureTfServing")
	tfServingStruct := GetTfServingConfig()
	tfServingStructPtr := &tfServingStruct
	if (*tfServingStructPtr).InstallIstio {
		installIstio()
	}

	emoji.Println(":two: Generating the service(model) components.")
	serviceComps := map[string]string{
		"modelName":   (*tfServingStructPtr).ModelName,
		"trafficRule": "v1:100",
		"serviceType": (*tfServingStructPtr).ServiceType,
	}

	fmt.Println("Setting service name")
	ksGenerateService := exec.Command("ks", "generate", "tf-serving-service", (*tfServingStructPtr).ServiceName)
	ksGenerateService.Dir = ksAppDir
	ksGenerateService.Stdin = os.Stdin
	ksGenerateService.Stdout = os.Stdout
	ksGenerateService.Stderr = os.Stderr

	err := ksGenerateService.Run()
	if err != nil {
		log.Fatal(Red(err))
	}

	for key, value := range serviceComps {
		switch key {
		case "modelName":
			fmt.Println("Setting deployment model name.")
			ksSetDeployName := exec.Command("ks", "param", "set", (*tfServingStructPtr).ServiceName, key, value)
			ksSetDeployName.Dir = ksAppDir
			ksSetDeployName.Stdin = os.Stdin
			ksSetDeployName.Stdout = os.Stdout
			ksSetDeployName.Stderr = os.Stderr
			err := ksSetDeployName.Run()
			if err != nil {
				log.Fatal(Red(err))
			}

		case "trafficRule":
			fmt.Println("Setting default traffic rule - v1:100.")
			ksSetTrafficRule := exec.Command("ks", "param", "set", (*tfServingStructPtr).ServiceName, key, value)
			ksSetTrafficRule.Dir = ksAppDir
			ksSetTrafficRule.Stdin = os.Stdin
			ksSetTrafficRule.Stdout = os.Stdout
			ksSetTrafficRule.Stderr = os.Stderr
			err := ksSetTrafficRule.Run()
			if err != nil {
				log.Fatal(Red(err))
			}

		case "serviceType":
			fmt.Println("Setting serviceType.")
			ksSetServiceType := exec.Command("ks", "param", "set", (*tfServingStructPtr).ServiceName, key, value)
			ksSetServiceType.Dir = ksAppDir
			ksSetServiceType.Stdin = os.Stdin
			ksSetServiceType.Stdout = os.Stdout
			ksSetServiceType.Stderr = os.Stderr
			err := ksSetServiceType.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		}
	}

	emoji.Println(":three: Generating the deployment(version) components.")
	fmt.Println("Setting MODEL_COMPONENT environment variable.")
	modelCompEnvVar := serviceComps["modelName"] + "-" + (*tfServingStructPtr).VersionName
	err = os.Setenv("$MODEL_COMPONENT", modelCompEnvVar)
	if err != nil {
		log.Fatal(Red(err))
	}

	deploymentComps := map[string]string{
		"modelName":     (*tfServingStructPtr).ModelName,
		"versionName":   (*tfServingStructPtr).VersionName,
		"modelBasePath": (*tfServingStructPtr).ModelBasePath,
	}

	fmt.Println("Generating deployment")
	fmt.Println(os.Getenv("$MODEL_COMPONENT"))
	ksGenerateDeploy := exec.Command("ks", "generate", (*tfServingStructPtr).DeploymentName, os.Getenv("$MODEL_COMPONENT"))
	ksGenerateDeploy.Dir = ksAppDir
	ksGenerateDeploy.Stdin = os.Stdin
	ksGenerateDeploy.Stdout = os.Stdout
	ksGenerateDeploy.Stderr = os.Stderr
	err = ksGenerateDeploy.Run()
	if err != nil {
		log.Fatal(Red(err))
	}

	for key, value := range deploymentComps {
		switch key {
		case "modelName":
			fmt.Println("Setting modelName for deployment component")
			ksSetModelName := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), key, value)
			ksSetModelName.Dir = ksAppDir
			ksSetModelName.Stdin = os.Stdin
			ksSetModelName.Stdout = os.Stdout
			ksSetModelName.Stderr = os.Stderr
			err := ksSetModelName.Run()
			if err != nil {
				log.Fatal(Red(err))
			}

		case "versionName":
			fmt.Println("Setting versionName for deployment component")
			ksSetVersionName := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), key, value)
			ksSetVersionName.Dir = ksAppDir
			ksSetVersionName.Stdin = os.Stdin
			ksSetVersionName.Stdout = os.Stdout
			ksSetVersionName.Stderr = os.Stderr
			err := ksSetVersionName.Run()
			if err != nil {
				log.Fatal(Red(err))
			}

		case "modelBasePath":
			fmt.Println("Setting modelBasePath for deployment component")
			fmt.Println("key", key)
			fmt.Println("value", value)
			ksSetModelBasepath := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), key, value)
			ksSetModelBasepath.Dir = ksAppDir
			ksSetModelBasepath.Stdin = os.Stdin
			ksSetModelBasepath.Stdout = os.Stdout
			ksSetModelBasepath.Stderr = os.Stderr
			err := ksSetModelBasepath.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		}
	}

	if (*tfServingStructPtr).NumGpus >= 1 {
		fmt.Println("Setting GPU parameter")
		ksSetGpuParam := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), "numGpus", strconv.Itoa((*tfServingStructPtr).NumGpus))
		ksSetGpuParam.Dir = ksAppDir
		ksSetGpuParam.Stdin = os.Stdin
		ksSetGpuParam.Stdout = os.Stdout
		ksSetGpuParam.Stderr = os.Stderr
		err := ksSetGpuParam.Run()
		if err != nil {
			log.Fatal(Red(err))
		}
	}

	if (*tfServingStructPtr).ModelLocation == "gcp" {
		fmt.Println("Setting GCP credentials secret name")
		ksSetSecretName := exec.Command("ks", "param", "set", "gcpCredentialSecretName", (*tfServingStructPtr).GcpSecretName)
		ksSetSecretName.Dir = ksAppDir
		ksSetSecretName.Stdin = os.Stdin
		ksSetSecretName.Stdout = os.Stdout
		ksSetSecretName.Stderr = os.Stderr
		err := ksSetSecretName.Run()
		if err != nil {
			log.Fatal(Red(err))
		}

	}
	var secretName string
	if (*tfServingStructPtr).ModelLocation == "s3" {
		if (*tfServingStructPtr).AwsSecretname == "" {
			fmt.Println("AWS secret name is not set. Creating the kubernetes secret")
			rand.Seed(time.Now().UnixNano())
			randomStr := RandStringBytes(5)
			secretName = "kf-tf-serving-secret-" + randomStr
			createSecretGeneric := exec.Command("kubectl", "-n", "kubeflow", "create", "secret", "generic",
				secretName, "--from-literal=AWS_ACCESS_KEY_ID="+os.Getenv("AWS_ACCESS_KEY_ID"),
				"--from-literal=AWS_SECRET_ACCESS_KEY="+os.Getenv("AWS_SECRET_ACCESS_KEY"))
			createSecretGeneric.Stdin = os.Stdin
			createSecretGeneric.Stdout = os.Stdout
			createSecretGeneric.Stderr = os.Stderr
			err := createSecretGeneric.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		} else {
			secretName = (*tfServingStructPtr).AwsSecretname
		}

		s3Params := map[string]string{
			"s3Enable":     strconv.FormatBool((*tfServingStructPtr).S3Enable),
			"s3SecretName": secretName,
		}

		for key, value := range s3Params {
			switch key {
			case "s3Enable":
				fmt.Println("Setting S3 related options for deployment")
				ksEnableS3 := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), key, value)
				ksEnableS3.Dir = ksAppDir
				ksEnableS3.Stdin = os.Stdin
				ksEnableS3.Stdout = os.Stdout
				ksEnableS3.Stderr = os.Stderr

				err := ksEnableS3.Run()
				if err != nil {
					log.Fatal(Red(err))
				}
			case "s3SecretName":
				fmt.Println("Setting S3 secret name.")
				ksSetSecretName := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), key, value)
				ksSetSecretName.Dir = ksAppDir
				ksSetSecretName.Stdin = os.Stdin
				ksSetSecretName.Stdout = os.Stdout
				ksSetSecretName.Stderr = os.Stderr

				err := ksSetSecretName.Run()
				if err != nil {
					log.Fatal(Red(err))
				}
			}
		}

		if len(strings.TrimSpace((*tfServingStructPtr).S3AwsRegion)) != 0 {
			fmt.Println("Setting AWS region")
			fmt.Println((*tfServingStructPtr).S3AwsRegion)
			ksSetS3Region := exec.Command("ks", "param", "set", "s3AwsRegion", (*tfServingStructPtr).S3AwsRegion)
			ksSetS3Region.Dir = ksAppDir
			ksSetS3Region.Stdin = os.Stdin
			ksSetS3Region.Stdout = os.Stdout
			ksSetS3Region.Stderr = os.Stderr

			err := ksSetS3Region.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		}

		if (*tfServingStructPtr).S3UseHttps {
			fmt.Println("Setting s3UseHttps option")
			ksSetS3Https := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), "s3UseHttps", strconv.FormatBool((*tfServingStructPtr).S3UseHttps))
			ksSetS3Https.Dir = ksAppDir
			ksSetS3Https.Stdin = os.Stdin
			ksSetS3Https.Stdout = os.Stdout
			ksSetS3Https.Stderr = os.Stderr

			err := ksSetS3Https.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		}

		if (*tfServingStructPtr).S3VerifySslCerts {
			fmt.Println("Setting s3VerifySsl option.")
			ksSetS3Verify := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), "s3VerifySsl", strconv.FormatBool((*tfServingStructPtr).S3VerifySslCerts))
			ksSetS3Verify.Dir = ksAppDir
			ksSetS3Verify.Stdin = os.Stdin
			ksSetS3Verify.Stdout = os.Stdout
			ksSetS3Verify.Stderr = os.Stderr

			err := ksSetS3Verify.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		}

		if (*tfServingStructPtr).S3EndpointUrl != "" {
			fmt.Println("Setting S3 endpoint URL")
			ksSetS3EpUrl := exec.Command("ks", "param", "set", os.Getenv("$MODEL_COMPONENT"), "s3Endpoint", (*tfServingStructPtr).S3EndpointUrl)
			ksSetS3EpUrl.Dir = ksAppDir
			ksSetS3EpUrl.Stdin = os.Stdin
			ksSetS3EpUrl.Stdout = os.Stdout
			ksSetS3EpUrl.Stderr = os.Stderr

			err := ksSetS3EpUrl.Run()
			if err != nil {
				log.Fatal(Red(err))
			}
		}

	}

	emoji.Println(":four: Applying the parameters for deployment")
	ksApplySvc := exec.Command("ks", "apply", os.Getenv("KF_ENV"), "-c", (*tfServingStructPtr).ServiceName)
	ksApplySvc.Dir = ksAppDir
	ksApplySvc.Stdin = os.Stdin
	ksApplySvc.Stdout = os.Stdout
	ksApplySvc.Stderr = os.Stderr

	err = ksApplySvc.Run()
	if err != nil {
		log.Fatal(Red(err))
	}

	ksApplyModelComp := exec.Command("ks", "apply", os.Getenv("KF_ENV"), "-c", os.Getenv("$MODEL_COMPONENT"))
	ksApplyModelComp.Dir = ksAppDir
	ksApplyModelComp.Stdin = os.Stdin
	ksApplyModelComp.Stdout = os.Stdout
	ksApplyModelComp.Stderr = os.Stderr

	err = ksApplyModelComp.Run()
	if err != nil {
		log.Fatal(Red(err))
	}
	emoji.Println(Green(":fire: TensorFlow model has been deployed successfully"))
}

func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func installIstio() {
	emoji.Println(":one: Installing istio")
	istioUrls := []string{"https://raw.githubusercontent.com/kubeflow/kubeflow/master/deployment/existing/istio/crds.yaml",
		"https://raw.githubusercontent.com/kubeflow/kubeflow/master/deployment/existing/istio/istio-noauth.yaml"}

	for _, element := range istioUrls {
		fmt.Println("element", element)
		applyIstio := exec.Command("kubectl", "apply", "-f", element)
		applyIstio.Stdin = os.Stdin
		applyIstio.Stdout = os.Stdout
		applyIstio.Stderr = os.Stderr

		err := applyIstio.Run()
		if err != nil {
			log.Fatal(Red(err))
		}

	}
	setIstioLabel := exec.Command("kubectl", "label", "namespace", "kubeflow", "istio-injection=enabled")
	setIstioLabel.Stdin = os.Stdin
	setIstioLabel.Stdout = os.Stdout
	setIstioLabel.Stderr = os.Stderr

	err := setIstioLabel.Run()
	if err != nil {
		log.Fatal(Red(err))
	}

	emoji.Println(Green(":white_check_mark: Istio has been deployed successfully"))
}
