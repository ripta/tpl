VERSION=$(shell git describe | sed -e 's/-g[0-9a-f]*$$//' -e 's/-/./' -e 's/^v//')

build:
	docker build --build-arg TPL_BUILD_DATE=$$(date +%Y%m%d-%H%M%S) --build-arg TPL_VERSION=$(VERSION) -t ripta/tpl:v$(VERSION) -f Dockerfile .
	docker tag ripta/tpl:v$(VERSION) ripta/tpl:latest

test-container: Dockerfile.test
	docker build -q -t ripta/tpl:test -f Dockerfile.test .

test-once: test-container
	docker run --rm -v $$(pwd):/go/src/github.com/ripta/tpl ripta/tpl:test \
		go test github.com/ripta/tpl

test-watch: test-container
	docker run -it --rm -v $$(pwd):/go/src/github.com/ripta/tpl ripta/tpl:test \
		looper

test: test-once

version:
	@echo $(VERSION)
