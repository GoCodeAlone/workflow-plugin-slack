.PHONY: build test install clean

BINARY_NAME = workflow-plugin-slack
INSTALL_DIR ?= data/plugins/$(BINARY_NAME)

build:
	go build -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

test:
	go test ./... -v -race

install: build
	mkdir -p $(DESTDIR)/$(INSTALL_DIR)
	cp bin/$(BINARY_NAME) $(DESTDIR)/$(INSTALL_DIR)/
	cp plugin.json $(DESTDIR)/$(INSTALL_DIR)/

clean:
	rm -rf bin/
