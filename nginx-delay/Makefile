IMAGE_NAME = splicemaahs/nginx-delay
VERSION = 0.0.1

.PHONY: all build-external clean-external push-external

all: build

build:
	docker build --rm -t $(IMAGE_NAME):$(VERSION) -f Dockerfile .
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest

push:
	docker push $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):$(VERSION)

clean:
	docker rmi $(IMAGE_NAME):latest
	docker rmi $(IMAGE_NAME):$(VERSION)

realclean:
	-docker kill $(shell docker ps -q)
	-docker rm $(shell docker ps -a -q)
	-docker rmi $(shell docker images -q)
	-docker system prune -f -a
