# Variables:
DIST_DIR = swiftwave_service/dashboard/www
SUBMOD_DIR = dashboard

main: build

deps:
	git submodule update --init && cd $(SUBMOD_DIR) && npm i

build_dashboard: | deps
	cd $(SUBMOD_DIR) &&	npm run build:swiftwave
	
clean_mkdir:
	rm -rf $(DIST_DIR) || true && \
	mkdir -p $(DIST_DIR)
	
copy_dashboard: | clean_mkdir build_dashboard
	cp -r $(SUBMOD_DIR)/dist/* $(DIST_DIR)
	
build: | copy_dashboard
	go build .
	
install: build
	cp swiftwave /usr/bin/swiftwave
