NAME = $(shell appv name)
VERSION = $(shell appv version)
IMAGE = $(shell appv image)

test:
	bash run-tests.sh

build:
	docker build -t $(IMAGE) .

delete:
	docker rmi $(IMAGE)

build-test:
	docker build -t "$(NAME)-test:$(VERSION)" -f Dockerfile.test .

run-docker-test:
	docker run --rm -t $(NAME)-test:$(VERSION)