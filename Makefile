
build:
	docker build -t ripta/tpl:latest -f Dockerfile .

test-container: Dockerfile.test
	docker build -q -t ripta/tpl:test -f Dockerfile.test .

test-once: test-container
	docker run --rm -v $$(pwd):/go/src/github.com/ripta/tpl ripta/tpl:test \
		go test github.com/ripta/tpl

test-watch: test-container
	docker run -it --rm -v $$(pwd):/go/src/github.com/ripta/tpl ripta/tpl:test \
		looper

test: test-once
