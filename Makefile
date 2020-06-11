TAG := $(shell git describe --tags)

.PHONY: unit-test
unit-test:
	cd pkg/reconciler && \
	go test

e2e-test:
	cd e2e && \
	bash test.sh

image-release:
	operator-sdk build ckotzbauer/access-manager:$(TAG) --go-build-args "-ldflags -X=main.Version=$(TAG)"
	docker tag ckotzbauer/access-manager:$(TAG) ckotzbauer/access-manager:latest
	docker push ckotzbauer/access-manager:$(TAG)
	docker push ckotzbauer/access-manager:latest
