TAG?=master
TRAVIS_COMMIT?=master
BINARY?=etcd-snapper-b2

build:
	go build -mod vendor -o $(BINARY)

unittest:
	go test ./... --run UnitTest

image:
	docker image build -t thomasjpfan/$(BINARY):$(TAG) \
	--label "org.opencontainers.image.revision=$(TRAVIS_COMMIT)" .
