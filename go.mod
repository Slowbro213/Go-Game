module gametry.com/main

go 1.24.3

require (
	gametry.com/handlers v0.0.0-00010101000000-000000000000
	gametry.com/player v0.0.0-00010101000000-000000000000
	github.com/gorilla/sessions v1.4.0
	github.com/gorilla/websocket v1.5.3
)

require github.com/gorilla/securecookie v1.1.2 // indirect

replace gametry.com/player => ./player

replace gametry.com/handlers => ./handlers
