VERSION=$(shell git describe | sed -e 's/-g[0-9a-f]*$$//' -e 's/-/./' -e 's/^v//')

build: build-docker build-local

build-docker:
	docker build --build-arg TPL_BUILD_DATE=$$(date +%Y%m%d-%H%M%S) --build-arg TPL_VERSION=$(VERSION) -t ripta/tpl:v$(VERSION) -f Dockerfile .
	docker tag ripta/tpl:v$(VERSION) ripta/tpl:latest

build-local:
	go install -ldflags "-s -w -X main.BuildVersion=$(VERSION) -X main.BuildDate=$$(date +%Y%m%d-%H%M%S)" .

push:
	git push
	git push --tags
	docker push ripta/tpl:v$(VERSION)
	docker push ripta/tpl:latest

test:
	go test ./...

version:
	@echo $(VERSION)
