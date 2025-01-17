clean :
	-rm -r build

build-content:
	$(MAKE) -C content build

fmt :
	$(info Reformatting all source files...)
	go fmt ./...

build : clean fmt build-content
	go build -o ./build/mullvad-web-controller go.roman.zone/mullvad-web-controller

install: build
	sudo ./service.sh install

uninstall:
	sudo ./service.sh uninstall

run-dev : build
	./build/mullvad-web-controller --dev --host localhost
