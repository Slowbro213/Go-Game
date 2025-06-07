
package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"gametry.com/handlers"
	"github.com/gorilla/sessions"
	"os"
)

var upgrader = websocket.Upgrader{}


func main() {

	//Static file serving
	styles := http.FileServer(http.Dir("./assets/css"))
	scripts := http.FileServer(http.Dir("./assets/js"))
	views := http.FileServer(http.Dir("./views"))
	http.Handle("/assets/css/", http.StripPrefix("/assets/css/", styles))
	http.Handle("/assets/js/", http.StripPrefix("/assets/js/", scripts))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w,r,"./assets/icons8-spartan-helmet-16.png")
	})
	http.Handle("/",http.StripPrefix("/",views))




	//Auth Services
	key   := []byte("super-secret-key-12345678") // 16, 24, or 32 bytes
	store := sessions.NewCookieStore(key)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure: true, // Enable this in production
	}
	l := log.New(os.Stdout, "SessionHandler: ", log.LstdFlags)
	sh := handlers.NewSessionHandler(store,l)
	http.HandleFunc("/secret", sh.Secret)
	http.HandleFunc("/login", sh.Login)
	http.HandleFunc("/logout", sh.Logout)


	l = log.New(os.Stdout, "ErrorHandler: ", log.LstdFlags)
	eh := handlers.NewErrorHandler(l)
	http.HandleFunc("/error/duplicate", eh.Duplicate)
	http.HandleFunc("/error/unauth", eh.UnAuthenticated)

	//Game WebSocket Service
	l = log.New(os.Stdout, "GameHandler: ", log.LstdFlags)
	gh := handlers.NewGameHandler(l,&upgrader,store)
	http.HandleFunc("/game", gh.Match )




	http.HandleFunc("/game/auth", gh.Auth)
	//Server Start
	fmt.Println("Server running at http://localhost:8080/")
	http.ListenAndServe("0.0.0.0:8080", nil)
}
