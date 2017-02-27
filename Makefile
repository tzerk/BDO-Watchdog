TARGET="BDO-Watchdog.exe"
VERSION=1.0.0
LDFLAGS=-ldflags="-H windowsgui"

# Dependencies:

all:
	go build ${LDFLAGS} -o ${TARGET} main.go
	