### Examples:

* Setup kubeflow:

  ```
  tk8ml install kubeflow --k8s
  ```      
  
* Remove kubeflow:
    * Preserve storage, remove everything:
    
      ```
      tk8ml delete kubeflow --all
      ```    

    * Delete Kubeflow along with storage

      ```
      tk8ml delete kubeflow --delete_storage
      ```  

* Install Kubeflow components:

    * Install chainer operator:  
        ```
        tk8ml install kubeflow-component --chainer-operator
        ```
    
    * Install katib (Hyperparameter Tuning)
    
        ```
        tk8ml install kubeflow-component --katib
        ```
        
    * Install ModelDB:
   
        ```
        tk8ml install kubeflow-component --modeldb
        ```
        
    * Install Seldon:         
         ```
         tk8ml install kubeflow-component --modeldb
         ```
     
* Setup kubeflow serving components:

    * Setup TensorFlow Batch Predict:
    
        * Make sure you have `config.yaml` in the directory. Example configuration for Tensorflow Batch Predict:
        ```yaml
        kf-components:
          serving:
            tf-batch-predict:
              kubeflow-version: "v0.6.1"
              registry-name: "kubeflow-git"
              job-name: "test-job-name"
              gcp-secret-name: ""
              input-file-patterns: "test-file-patterns"
              input-file-format: "json"
              model-path: "test-model-path"
              batch-size: 3
              output-result-prefix: "test-op-prefix"
              output-error-prefix: "test-error-prefix"
              num-gpu:
        ```    
        where:
        
        * `kubeflow-version`: The kubeflow version which you want to use while cloning the Kubeflow repository.
        * `registry-name`: The registry name which you want to use while cloning the Kubeflow repository.
        * `job-name`: The name of the tensorflow batch predict job.
        * `gcp-secret-name`: Secret name if used on GCP. Only needed for running the jobs in GKE in order to output results to GCS.
        * `input-file-patterns`: The list of input files or file patterns, separated by commas.
        * `input-file-format`: One of the following values: json, tfrecord, and tfrecord_gzip.
        * `model-path`: The path containing the model files in SavedModel format.
        * `batch-size`: Number of prediction instances in one batch. This largely depends on how many instances can be held and processed simultaneously in the memory of your machine.
        * `output-result-prefix`: Output path to save the prediction results.
        * `output-error-prefix`: Output path to save the prediction errors.
        * `num-gpu`: Number of GPUs to use per machine.
                     
        * Once `config.yaml` is ready, run:
            ```
            tk8ml install kubeflow-serving --tf-batch-predic
            ```
            
    * Setup Tensorflow Serving:
        * `export` AWS related values in the environment:
        
        ```
        export AWS_ACCESS_KEY_ID=XXXXXXXXXXX
        export AWS_SECRET_ACCESS_KEY=XXXXXXXXXXXXXXXXXXX
        ```
        * Make sure you have `config.yaml` in the directory. Example configuration for Tensorflow Serving:
        ```yaml
        kf-components:
          serving:
            tf-serving:
              install-istio: false
              tf-serving-service-name: "mnist-test"
              tf-serving-deployment-name: "tf-serving-deployment-aws"
              model-name: "mnist-model-name"
              service-type: "LoadBalancer"
              version-name: "v1"
              model-base-path: "s3://kubeflow-models/inception"
              gcp-secret-name:
              aws-secret-name:
              inject-istio: true
              s3-enable: true
              s3-secret-name:
              s3-aws-region:
              s3-use-https: true
              s3-verify-ssl-certs: true
              s3-endpoint-url: "s3.eu-central-1.amazonaws.com"
              num-gpus: 0
              model-location: s3 # supported are gcp/s3
        ```    
        where:
                
         * `install-istio`: Install istio with the model.
         * `tf-serving-service-name`: The service name.
         * `tf-serving-deployment-name`: The deployment name.
         * `model-name`: TF serving model name.
         * `service-type`: The service type for the TF serving.
         * `version-name`: This value will be set for `versionName` for this component.
         * `model-base-path`: The path where the model is present.
         * `gcp-secret-name`: To be used if the model is to be deployed on GKE.
         * `aws-secret-name`: To be used if the model is deployed on Kubernetes cluster built on AWS. If not set, the secret will automatically be created. The auto-created secret will contain the values of `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` (of course they will be base64 encoded).
         * `inject-istio`: Set this to true if you want to inject istio in TF serving.
         * `s3-enable`: This will set `s3Enable` parameter. 
         * `s3-secret-name`: Set this to the same value of `aws-secret-name`, if the secret already exists. In case of auto-generated secret, keep this field empty (this will be done internally).
         * `s3-aws-region`: (Optional) Set AWS region for S3.
         * `s3-use-https`: (Optional) Whether or not to use https for S3 connections.
         * `s3-verify-ssl-certs`: (Optional) Whether or not to verify https certificates for S3 connections.
         * `s3-endpoint-url`: (Optional) URL for your s3-compatible endpoint.
         * `num-gpus`: (Optional) If you want to use GPU to serve the model.
         * `model-location`: The location where the model is location. Supported values are `gcp` and `s3`.
         
         * Once `config.yaml` is ready, run:
            ```
            tk8ml install kubeflow-serving --tf-serving
            ``` 