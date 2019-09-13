# Pure1 Unplugged
Pure1 Unplugged is Pure Storage's open-source on-premises fleet management system.

## Installation
Pure1 Unplugged can be installed from an ISO or an OVA. The latest versions, along with documentation, are pre-created and are available to download for free. Documentation is also available online.

### Links
http://dl-west.purestorage.com/pure/pure1-unplugged/download

https://support.purestorage.com/Pure1/Pure1_Unplugged.

## Development
Pure1 Unplugged can be built and run locally.

### Prerequisites
1. Forked and cloned source code
2. Minikube and required hypervisor: https://kubernetes.io/docs/tasks/tools/install-minikube/

### Building
There are two separate Makefiles for building and testing. The main [Makefile](./Makefile) builds all available targets. The sub-Makefile 
[Makefile-golang](./Makefile-golang) has targets for building golang in a docker container. The top level Makefile looks for any target that matches the pattern `go-%` and will launch a container and make inside the container with the targets matching pattern.

_Example_: `make go-clean go-prep go-auth-server` runs the targets `clean prep auth-server` in a container.

### Testing
* Golang: `make go-unit-tests`
* Web content: `make test-web-content`

### Local deployment
1. Start minikube with at least 6GB memory and 40GB disk
   *  `minikube start --memory 6144 --disk-size 40g`
2. Build with `make all-minikube`
    * Note that build output goes into `./build/*`, and this `minikube` build mode will *not* create the installer ISO
3. Deploy with `./scripts/deploy/helm_install.sh $(minikube ip)`
    * Note that the deployment name must be `pure1-unplugged`
4. Run `minikube ip` and navigate to it in your browser
5. Log in with example
 
_Hint:_ Set `kubectl` to use `pure1-unplugged` namespace by running `kubectl config set-context $(kubectl config current-context) --namespace=pure1-unplugged` for easier use.

### Updating a deployment

#### Updating golang
```
make go-bins
make pure-elk-image-minikube
make lorax-image
make helm-chart
helm upgrade pure-elk build/chart/*.tgz --force
```

#### Updating web content
```
make web-content
make pure-elk-image-minikube
make lorax-image
make helm-chart
helm upgrade pure-elk build/chart/*.tgz --force
```

## Contributions

### Changes
As Pure1 Unplugged is open-source, we welcome all contributions. We will also be providing improvements and new features.

1. Create a public fork of the repository
2. Send a pull request (we need to create a reviewer group for us)
3. We will test and validate your code
4. Merge the pull request
5. We will update the ISO and OVA to download and comment on the pull request

### Issues
We will also monitor issues listed on GitHub, as well as create issues ourselves. All changes should reference an issue.

