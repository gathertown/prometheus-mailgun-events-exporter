buildlinux:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/mailgun_events_exporter-linux ./cmd/*

darwinbuild:
	CGO_ENABLED=0 GOOS=darwin go build -o ./bin/mailgun_events_exporter-darwin ./cmd/*

build: buildlinux darwinbuild

build-image:
	docker build -t gathertown/prometheus-mailgun-events-exporter:latest .

run: build
	go run ./cmd/main.go

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -rf bin/
