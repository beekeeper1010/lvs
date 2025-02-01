BUILD     := build/
BIN       := lvs2.exe
TARGET    := $(BUILD)$(BIN)
RESOURCES := $(wildcard web/ *.db *.yaml)

$(shell mkdir -p $(BUILD))

.PHONY: all clean

all: $(TARGET)
	@echo completed!

$(TARGET):
	buildTime=`date +%FT%T` && \
	commitId=`git log -1 --pretty=%h` && \
	go build -ldflags="-s -w -X main.buildTime=$$buildTime -X main.commitId=$$commitId" -o $@ && \
	upx -9 $@ && \
	cp -rf $(RESOURCES) $(BUILD)

clean:
	rm -rf $(BUILD)