# Copyright 2021 ABSA Group Limited
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
# The binary to build (just the basename).
BIN := k8s_crd

# Where to push the docker image.
REGISTRY ?= docker.io/absaoss

# Tag 
TAG ?= latest

# Image URL to use all building/pushing image targets
IMG ?= $(REGISTRY)/$(BIN)

# find or download golic
# download golic if necessary
golic:
ifeq (, $(shell which golic))
	@{ \
	set -e ;\
	GOLIC_TMP_DIR=$$(mktemp -d) ;\
	cd $$GOLIC_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/AbsaOSS/golic@v0.4.1 ;\
	rm -rf $$GOLIC_TMP_DIR ;\
	}
GOLIC=$(GOBIN)/golic
else
GOLIC=$(shell which golic)
endif

# run all linters from .golangci.yaml; see: https://golangci-lint.run/usage/install/#local-installation
.PHONY: lint
lint:
	golangci-lint run

build:
	GOOS=linux CGO_ENABLED=0 go build cmd/coredns.go

clean:
	go clean
	rm -f coredns

image: 
	docker build . -t ${IMG}:${TAG}

create-local-cluster:
	k3d cluster create -c k3d-cluster.yaml

import-image:
	k3d image import -c coredns-crd absaoss/k8s_crd:${TAG}

deploy-app: image import-image
	kubectl config use-context k3d-coredns-crd
	kubectl apply -f terratest/example/ns.yaml 
	kubectl create -n coredns configmap geodata --from-file terratest/geogen/geoip.mmdb || true
	kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/external-dns/master/docs/contributing/crd-source/crd-manifest.yaml
	helm repo add coredns https://coredns.github.io/helm
	helm repo update
	cd charts/coredns && helm dependency update
	helm upgrade -i coredns -n coredns charts/coredns \
		-f terratest/helm_values.yaml \
		--set coredns.image.tag=${TAG}

.PHONY: lincense
# updates source code with license headers
license: golic
	$(GOLIC) inject -c "2021 ABSA Group Limited"

.PHONY: terratest
terratest: deploy-app
	cd terratest/test/ && go test -v

