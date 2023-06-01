module github.com/lucasl0st/InfiniteDB/tools/idbdump

go 1.20

replace github.com/lucasl0st/InfiniteDB => ../../

replace github.com/lucasl0st/InfiniteDB/idblib => ../../idblib

replace github.com/lucasl0st/InfiniteDB/tools/util => ../util

replace github.com/lucasl0st/InfiniteDB/server => ../../server

require (
	github.com/lucasl0st/InfiniteDB v0.0.0-00010101000000-000000000000
	github.com/lucasl0st/InfiniteDB/idblib v0.0.0-20230420200119-e5e17987ddf0
	github.com/lucasl0st/InfiniteDB/server v0.0.0-00010101000000-000000000000
	github.com/lucasl0st/InfiniteDB/tools/util v0.0.0-00010101000000-000000000000
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/spf13/cobra v1.7.0
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gammazero/deque v0.2.0 // indirect
	github.com/gammazero/workerpool v1.1.3 // indirect
	github.com/gin-gonic/gin v1.9.1 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)
