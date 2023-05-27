module github.com/lucasl0st/InfiniteDB/integration_tests

go 1.20

replace github.com/lucasl0st/InfiniteDB => ../

replace github.com/lucasl0st/InfiniteDB/idblib => ../idblib

require (
	github.com/dimchansky/utfbom v1.1.1
	github.com/lucasl0st/InfiniteDB v0.0.0-00010101000000-000000000000
	github.com/lucasl0st/InfiniteDB/idblib v0.0.0-00010101000000-000000000000
	vimagination.zapto.org/dos2unix v1.0.0
)

require (
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)
