
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"log"
	"game/handlers"
	"game/middleware"
	"github.com/gorilla/sessions"
	"os"
	//"time"
	//"runtime"
)



func main() {

	//Static file serving
	styles := http.FileServer(http.Dir("./assets/css"))
	scripts := http.FileServer(http.Dir("./assets/js"))
	wasm   := http.FileServer(http.Dir("./wasm"))
	errors := http.FileServer(http.Dir("./views/errors"))
	http.Handle("/assets/css/", http.StripPrefix("/assets/css/", styles))
	http.Handle("/assets/js/", http.StripPrefix("/assets/js/", scripts))
	http.Handle("/wasm/", http.StripPrefix("/wasm/", wasm))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w,r,"./assets/icons8-spartan-helmet-16.png")
	})
	http.Handle("/error",http.StripPrefix("/",errors))





	//Auth Services
	key   := []byte("super-secret-key-12345678") // 16, 24, or 32 bytes
	store := sessions.NewCookieStore(key)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure: true, // Enable this in production
	}


	authService := middleware.NewAuthService(store, log.New(os.Stdout, "AUTH: ", log.LstdFlags))


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
	gh := handlers.NewGameHandler(l,store)
	http.HandleFunc("/game", middleware.Chain(
		gh.Match,
		middleware.Logging(),
		authService.AuthMiddleware(),
		middleware.Method("GET"),
	))


	http.HandleFunc("/triangle", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w,r,"./views/triangle.html")
	})


	http.HandleFunc("/", middleware.Chain(
		gh.Join,
		middleware.Logging(),
		authService.AuthMiddleware(),
		middleware.Method("GET"),
		))

	//go func() {
	//	for {
	//		time.Sleep(2 * time.Second)
	//		runtime.GC()
	//	}
	//}()
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()
	//Server Start
	fmt.Println("Server running at http://localhost:8080/")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		panic("Couldnt start server")
	}
}
