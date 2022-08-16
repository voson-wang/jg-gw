module jg-gateway

go 1.19

require (
	e.coding.net/ricnsmart/service/jg-modbus v0.0.1
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/joho/godotenv v1.4.0
	github.com/rs/zerolog v1.27.0
	github.com/shopspring/decimal v1.3.1
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220808155132-1c4a2a72c664 // indirect
)

replace (
	e.coding.net/ricnsmart/service/jg-modbus v0.0.1 => ../jg-modbus
)