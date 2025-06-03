TARGETS := fileserver xseld
.PHONY: $(TARGETS)

all: $(TARGETS)

$(TARGETS):
	cd $@ && go test ./... && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../rootfs/usr/bin/$@ .
