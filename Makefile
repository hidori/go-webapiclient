AUTHOR = hidori
PROJECT = gontext

.PHONY: lint
lint:
	docker run --rm -v ${PWD}:${PWD} -w ${PWD} golangci/golangci-lint:latest-alpine golangci-lint run

.PHONY: format
format:
	docker run --rm -v ${PWD}:${PWD} -w ${PWD} golangci/golangci-lint:latest-alpine golangci-lint run --fix

.PHONY: test
test:
	go test -cover .
	go run ./example/example.go

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
