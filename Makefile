SHELL:=/bin/bash

.PHONY: all
all: clean build

.PHONY: build
build: cerbos-binary image publish-image publish-lambda update-lambda

.PHONY: clean
clean:
	@ rm -rf .cerbos

.PHONY: cerbos-binary
cerbos-binary:
	@ if [[ "$$CERBOS_RELEASE" ]]; then \
		CURRENT_RELEASE=$$CERBOS_RELEASE; \
	else \
		CURRENT_RELEASE=$$(curl -sH "Accept: application/vnd.github.v3+json"  https://api.github.com/repos/cerbos/cerbos/tags | grep -o -E '"name": "v\d+\.\d+.\d+"' | head -1); \
	fi; \
	ver='[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+'; \
	if [[ $$CURRENT_RELEASE =~ $$ver  ]]; then \
		CURRENT_RELEASE="$${BASH_REMATCH[0]}"; \
	else \
		echo "Unexpected format of CERBOS_RELEASE, expected semantic version 'x.x.x'" >&2; \
		exit 1; \
	fi; \
	arch=$$(uname -m); [ "$$arch" != "x86_64" ] && [ "$$arch" != "arm64" ] && { echo "$${arch} - unsupported architecture, supported: x86_64, arm64" >&2; exit 1; }; \
	echo "Downloading Cerbos binaries if necessary"; \
	for os in Linux Darwin; do \
		a=$$arch; \
		if [ "$$a" = "amd64" ]; then \
			a=x86_64; \
		fi; \
		p=$${os}_$${a}; \
		mkdir -p ./.cerbos/$${p}; \
		[ -e "./.cerbos/$${p}/cerbos" ] || curl -sL "https://github.com/cerbos/cerbos/releases/download/v$${CURRENT_RELEASE}/cerbos_$${CURRENT_RELEASE}_$${os}_$${a}.tar.gz" | tar -xz -C ./.cerbos/$${p}/ cerbos; \
    done; \

.PHONY: image
image: cerbos-binary
	@ arch=$$(uname -m); [ "$$arch" != "x86_64" ] && [ "$$arch" != "arm64" ] && { echo "$${arch} - unsupported architecture, supported: x86_64, arm64" >&2; exit 1; }; \
	docker build --build-arg ARCH=$$arch -t cerbos/aws-lambda-gateway .

.PHONY: ecr
ecr:
	@ [ -n "$$ECR_REPOSITORY_URL" ] || { echo "Please set ECR_REPOSITORY_URL environment variable" >&2; exit 1; }

.PHONY: publish-image
publish-image: image ecr
	docker tag  cerbos/aws-lambda-gateway:latest $${ECR_REPOSITORY_URL}:latest
	docker push $${ECR_REPOSITORY_URL}:latest

.PHONY: publish-lambda
publish-lambda: ecr
	@ arch=$$(uname -m); [ "$$arch" != "x86_64" ] && [ "$$arch" != "arm64" ] && { echo "$${arch} - unsupported architecture, supported: x86_64, arm64" >&2; exit 1; }; \
	sam deploy --template sam.yml --stack-name $${CERBOS_STACK_NAME:-Cerbos} --resolve-image-repos \
	 --capabilities CAPABILITY_IAM --no-confirm-changeset  --no-fail-on-empty-changeset --parameter-overrides ArchitectureParameter=$$arch

.PHONY: update-lambda
update-lambda: ecr
	fn=$$(aws cloudformation list-stack-resources --stack-name $${CERBOS_STACK_NAME:-Cerbos} --query "StackResourceSummaries[?ResourceType=='AWS::Lambda::Function'].PhysicalResourceId" --output text); \
	aws lambda update-function-code --function-name $$fn --image-uri $${ECR_REPOSITORY_URL}:latest > /dev/null; \
	aws lambda wait function-updated --function-name $$fn


.PHONY: test
test:
	go test -v ./...
