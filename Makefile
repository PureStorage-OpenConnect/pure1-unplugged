
all-rpm: clean go web pure1-unplugged-image lorax-image helm-chart install-bundle rpm

all: all-rpm iso

all-minikube: clean go web pure1-unplugged-image-minikube lorax-image helm-chart

clean:
	rm -rf ./build

go: pull-go-image go-all

pull-go-image:
	./scripts/pull_gobuilder.sh

# Match any target that starts with 'go-' and shim them into the docker container where we run
# the target there without the 'go-' prefix. Example: make go-auth-server will run (in docker)
# something like make -f ./Makefile-golang auth-server
go-%:
	./scripts/build/build_golang_in_docker.sh $*

web: pull-web-image web-setup lint-web-content test-web-content web-content

pull-web-image:
	./scripts/pull_angularbuilder.sh

web-setup:
	./scripts/setup_web_content_in_docker.sh

lint-web-content:
	./scripts/build/lint_web_content_in_docker.sh

web-content:
	./scripts/build/build_web_content_in_docker.sh

test-web-content:
	./scripts/build/test_web_content_in_docker.sh

pure1-unplugged-image:
	./scripts/build/build_pure1_unplugged_image.sh

pure1-unplugged-image-minikube:
	./scripts/build/build_pure1_unplugged_image.sh minikube

lorax-image:
	./scripts/build/build_lorax_image.sh

helm-chart:
	./scripts/build/build_helm_chart_in_docker.sh

install-bundle:
	./scripts/build/build_install_bundle_in_docker.sh

rpm:
	./scripts/build/build_pure1_unplugged_rpm_in_docker.sh

iso:
	./scripts/build/build_installer_iso_in_docker.sh

.PHONY: all-rpm all-minikube clean go pull-go-image web pull-web-image web-setup pure1-unplugged-image pure1-unplugged-image-minikube lorax-image helm-chart install-bundle rpm install-iso
