DOCKER_LINT_CMD = docker run --rm -v $(PWD):$(PWD) -w $(PWD) golangci/golangci-lint:latest-alpine

.PHONY: lint
lint:
	$(DOCKER_LINT_CMD) golangci-lint config verify
	$(DOCKER_LINT_CMD) golangci-lint run

.PHONY: format
format:
	$(DOCKER_LINT_CMD) golangci-lint config verify
	$(DOCKER_LINT_CMD) golangci-lint run --fix

.PHONY: test
test:
	go test -v -cover .

.PHONY: run
run:
	go run ./example/main.go

.PHONY: version/patch
version/patch: test lint
	git fetch
	git checkout main
	git pull
	docker run --rm hidori/semver -i patch `cat ./meta/version.txt` > ./meta/version.txt
	git add ./meta/version.txt
	git commit -m 'Updated version.txt'
	git push
	git tag v`cat ./meta/version.txt`
	git push origin --tags
