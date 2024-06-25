.PHONY: update
update:
	go get -u
	go vet
