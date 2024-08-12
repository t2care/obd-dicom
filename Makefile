scp:
	go run cmd/obd-dicom/main.go -scp -datastore . -port 1104 -calledae ANY-SCP

.PHONY: test
test:
	go test -p 1 -v ./...