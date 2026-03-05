buildlinux:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/mailgun_events_exporter-linux ./cmd/*

darwinbuild:
	CGO_ENABLED=0 GOOS=darwin go build -o ./bin/mailgun_events_exporter-darwin ./cmd/*

build: buildlinux darwinbuild

build-image:
	docker build -t gathertown/prometheus-mailgun-events-exporter:latest .

# Build and push multi-arch image (linux/amd64, linux/arm64). Requires: docker login
push-multi-arch:
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t gathertown/prometheus-mailgun-events-exporter:latest \
		--push .

run: build
	go run ./cmd/main.go

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -rf bin/
