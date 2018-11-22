NAME = $(shell appv name)
VERSION = $(shell appv version)
IMAGE = $(shell appv image)

test:
	sh run-tests.sh

build:
	docker build -t $(IMAGE) .

build-test:
	docker build -t "$(NAME)-test:$(VERSION)" -f Dockerfile.test .

run-docker-test:
	docker run --rm $(NAME)-test:$(VERSION)