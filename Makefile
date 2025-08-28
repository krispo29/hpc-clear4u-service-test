PROJECT_ID = the-deft-technology
IMAGE_NAME = hpc-express-service
version ?= v0.0.16
# version ?= staging
VERSION = ${version}
EXPOSE_PORT = 80
DOCKER_PORT = 6200

GCR_URL_STAGING = gcr.io/${PROJECT_ID}/${IMAGE_NAME}:staging
GCR_URL = gcr.io/${PROJECT_ID}/${IMAGE_NAME}:${VERSION}

.PHONY: build
build:
	docker build -t ${GCR_URL} .


.PHONY: push
push:
	docker -- push ${GCR_URL}

.PHONY: push-staging
push-staging:
	docker -- push ${GCR_URL_STAGING}


.PHONY: test
test:
	docker run --rm -v $(pwd)/.env:/.env -v $(pwd)/private.pem:/private.pem -v $(pwd)/public.pem:/public.pem -p ${DOCKER_PORT}:${DOCKER_PORT} ${GCR_URL}

.PHONY: prod
prod:
	docker run -d -p ${EXPOSE_PORT}:${DOCKER_PORT} ${GCR_URL}
	
.PHONY: staging
staging:
	docker build -t ${GCR_URL_STAGING} .
	docker -- push ${GCR_URL_STAGING}

.PHONY: all
all: build push