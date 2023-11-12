module goatcli

go 1.21

require github.com/go-ini/ini v1.67.0

require (
	github.com/google/uuid v1.4.0
	github.com/robert-kisteleki/goatapi v0.6.0
)

require (
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/miekg/dns v1.1.56 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
)

// uncomment the next line for local goatcli/goatapi development (adapt path as needed)
//replace github.com/robert-kisteleki/goatapi => ../goatapi
