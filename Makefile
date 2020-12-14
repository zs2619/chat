BACKEND_APP_NAME=kissrtm
Version=v1
DOCKER_BUILD_IMAGE_TAG=kissrtm-build
DOCKER_RUN_IMAGE_TAG=kissrtm-run

GitCount= $(shell git rev-list --count HEAD)
branch=$(shell git symbolic-ref --short HEAD)
Branch=$(subst /,_,$(branch))

.PHONY: build  buildimage test clean vet deployimage dockertest 

build:
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
			-f ./Dockerfile.build .
	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) > ${BACKEND_APP_NAME}_${Version}_${Branch}_${GitCount}.tar.gz
buildimage:
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
			-f ./Dockerfile.build .

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile.run -
	docker tag $(DOCKER_RUN_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_IMAGE_TAG):latest


test: export LOGLEVEL=info
test:
	go test  -cover -race -parallel 1 -p 1 -count=1 -v ./... | sed '/PASS/s//$(shell printf "\033[32mPASS\033[0m")/' | sed '/FAIL/s//$(shell printf "\033[31mFAIL\033[0m")/' | sed '/coverage/s//$(shell printf "\033[32mcoverage\033[0m")/'

dockertest:
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
		-f ./Dockerfile.build .
	docker-compose -f "docker-compose-test.yaml" up -d --build
	docker logs -f kissrtm-build-test 
	docker-compose -f "docker-compose-test.yaml" down

vet :
	go vet ./...

deployimage:
	make buildimage
	docker-compose -f "docker-compose-image.yaml" down
	docker-compose -f "docker-compose-image.yaml" up -d

clean:
