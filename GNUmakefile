PLATFORM ?= $(shell go env GOOS)_$(shell go env GOARCH)

default: testacc

install:
	go build -o ~/.terraform.d/plugins/registry.terraform.io/materializeinc/mzcloud/1.0.0/$(PLATFORM)/terraform-provider-mzcloud

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
