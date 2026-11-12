.PHONY : clean build-content fmt vet test build install uninstall run-dev

clean :
	-rm -r build

build-content :
	$(MAKE) -C content build

fmt :
	$(info Reformatting all source files...)
	go fmt ./...

vet :
	go vet ./...

test : fmt vet
	go test ./...

build : clean fmt build-content
	go build -o ./build/mullvad-web-controller go.roman.zone/mullvad-web-controller

install : build
	sudo ./service.sh install

uninstall :
	sudo ./service.sh uninstall

run-dev : build
	MULLVAD_WEB_DRY_RUN=1 MULLVAD_WEB_ADDR=localhost:8666 ./build/mullvad-web-controller
