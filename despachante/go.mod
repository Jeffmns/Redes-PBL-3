module pbl/despachante

go 1.25.8

replace pbl/core => ../core

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1
	pbl/core v0.0.0-00010101000000-000000000000
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
)
