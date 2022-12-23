GO_DIR ?= $(shell pwd)

build: build_ingress-controller build_api

build_api:
	go build ./cmd/api/

build_ingress-controller:
	go build ./cmd/ingress-controller/

build_docker: build_ingress-controller_docker

build_ingress-controller_docker:
	docker build --file ./build/ingress-controller.Dockerfile --tag mck8s-ingress-controller:latest .
