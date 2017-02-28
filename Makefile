TARGET="BDO-Watchdog.exe"
VERSION=1.0.0
LDFLAGS=-ldflags="-H windowsgui"

# Dependencies:
# https://github.com/electron/rcedit
# - the rcedit.exe must be in PATH

all:
	go build ${LDFLAGS} -o ${TARGET} main.go
	rcedit ${TARGET} --set-icon blackspirit.ico