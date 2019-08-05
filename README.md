## TK8ml - A CLI to deploy and run ML workflows with Kubeflow

TK8ml is a command-line tool written in Go. It automates the creation and deployment of machine learning workflows using [Kubeflow](https://github.com/kubeflow/kubeflow). With TK8ml, you can create and manage multiple [kubeflow components](https://www.kubeflow.org/docs/components/) by simple and manageable configuration files.

As of now, TK8ml uses Kubeflow components. In near future, if more options are available around running ML workloads on Kubernetes, TK8ml will add support for them as well.


### Prerequisites: 
* [kfctl binary](https://github.com/kubeflow/kubeflow/releases/).
* Configured aws CLI (if using AWS as a cloud provider).
* A Kubernetes cluster.

### Installation

#### Downloading binary from the releases
* Download the latest binary (platform-specific) from the [releases](https://github.com/kubernauts/tk8ml/releases).

#### Building the binary
* Clone the repository
* `cd` into the repository
* Run: `make bin`. This will build the binary with the name `tk8ml`.

### Kubeflow operations supported so far:
* Install/Remove Kubeflow on a Kubernetes cluster. EKS/AKS/GKE and other supported platforms will be added soon.
* Setup Kubeflow components:
    * [Chainer Operator](https://www.kubeflow.org/docs/components/training/chainer/)
    * [Katib](https://www.kubeflow.org/docs/components/hyperparameter-tuning/hyperparameter/)
    * ModelDB
    * [Seldon](https://www.kubeflow.org/docs/components/serving/seldon/)

* Kubeflow serving:
    * [TensorFlow Batch Predict](https://www.kubeflow.org/docs/components/serving/tfbatchpredict/)
    * [TensorFlow Serving](https://www.kubeflow.org/docs/components/serving/tfserving_new/)

### Contributing:
This project is in the initial stages. Contributions are always welcome. See [Issues](https://github.com/kubernauts/tk8ml/issues). Or if you feel that something needs to be added, feel free to open an issue.

You can also report the issue or initiate the discussion around `tk8ml` via Slack.

[Join us on Kubernauts Slack Channel](https://kubernauts-slack-join.herokuapp.com/)


### Credits:
* [Kubeflow](https://www.kubeflow.org)