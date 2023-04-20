module github.com/lucasl0st/InfiniteDB/tools/idbimport

go 1.20

replace github.com/lucasl0st/InfiniteDB => ../../

replace github.com/lucasl0st/InfiniteDB/models => ../../models

replace github.com/lucasl0st/InfiniteDB/tools/util => ../util

require (
	github.com/lucasl0st/InfiniteDB v0.0.0-00010101000000-000000000000
	github.com/lucasl0st/InfiniteDB/tools/util v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.7.0
)

require (
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/lucasl0st/InfiniteDB/models v0.0.0-00010101000000-000000000000 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/sys v0.6.0 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)
