.PHONY: release all test race cover cover-html clean travis

release:
	@echo "Release v$(version)"
	@git pull
	@git checkout master
	@git pull
	@git checkout develop
	@git flow release start $(version)
	@git flow release finish $(version) -p -m "Release v$(version)"
	@git checkout develop
	@echo "Release v$(version) finished."

all: coverage.out

coverage.out: $(shell find . -type f -print | grep -v vendor | grep "\.go")
	@go test -cover -coverprofile ./coverage.out.tmp ./...
	@cat ./coverage.out.tmp | grep -v '.pb.go' | grep -v 'mock_' > ./coverage.out
	@rm ./coverage.out.tmp

test: coverage.out

race: $(shell find . -type f -print | grep -v vendor | grep "\.go")
	@go test -race ./...

cover: coverage.out
	@echo ""
	@go tool cover -func ./coverage.out

cover-html: coverage.out
	@go tool cover -html=./coverage.out

clean:
	@rm ./coverage.out
	@go clean -i ./...

travis: race coverage.out
