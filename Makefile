# Variables:
DIST_DIR = swiftwave_service/dashboard/www
SUBMOD_DIR = dashboard

main: build_service

build_dashboard:
	npm run build:dashboard
	
build_service: | build_dashboard
	go build .
	
install: build_service
	cp swiftwave /usr/bin/swiftwave
