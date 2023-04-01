GOBIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --abbrev-ref HEAD)
COMMIT ?= $(shell git describe --match=NeVeRmAtCh --always --abbrev=40 --dirty)
BUILDFLAGS ?= -gcflags '$(GCFLAGS)' -ldflags '$(LDFLAGS) -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)' -tags '$(BUILD_TAGS)'
IMAGE_REPO ?= xhebox
IMAGE_TAG ?= latest

.PHONY: cmd_%

default: cmd_csi-rclone

cmd_%: OUTPUT=$(patsubst cmd_%,./bin/%,$@)
cmd_%: SOURCE=$(patsubst cmd_%,./cmd/%,$@)
cmd_%:
	go build $(BUILDFLAGS) -o $(OUTPUT) $(SOURCE)

docker-release:
	podman build --platform linux/amd64 -t "localhost/$(IMAGE_REPO)/csi-rclone:$(IMAGE_TAG)" --build-arg "GOPROXY=$(shell go env GOPROXY)" --build-arg "VERSION=$(VERSION)" --build-arg "COMMIT=$(COMMIT)" -f docker/Dockerfile .
	podman push "localhost/$(IMAGE_REPO)/csi-rclone:$(IMAGE_TAG)" docker://$(IMAGE_REPO)/csi-rclone:$(IMAGE_TAG)
