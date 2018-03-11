
TARGET=magsli
BIN_DIR=./bin

VERSION=`git describe`
GIT_HASH=`git rev-parse HEAD`
BUILDTIME=`date -u '+%Y-%m-%d %H:%M:%S'`
LD_FLAGS="-X \"main.version=$(VERSION)\" -X main.githash=$(GIT_HASH) -X \"main.buildstamp=$(BUILDTIME)\""

all: macos linux

macos:
	GOOS=darwin go build -ldflags $(LD_FLAGS) -o $(BIN_DIR)/macos/$(TARGET)

linux:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LD_FLAGS) -a -installsuffix cgo -o $(BIN_DIR)/linux/$(TARGET)

clean:
	rm -rf $(BIN_DIR)
