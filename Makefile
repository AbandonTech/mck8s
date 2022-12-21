GO_DIR ?= $(shell pwd)

build: build_ingress-controller

build_ingress-controller:
	go build cmd/ingress-controller.go

build_docker: build_ingress-controller_docker

build_ingress-controller_docker:
	docker build --file build/ingress-controller.Dockerfile --tag mck8s-ingress-controller:latest .
