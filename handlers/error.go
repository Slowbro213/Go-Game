package handlers

import (
	"log"	
	"net/http"
)

type ErrorHandler struct{
	log *log.Logger
}


func NewErrorHandler(l* log.Logger) * ErrorHandler{
	return &ErrorHandler{log:l}
}



func (e* ErrorHandler) Duplicate(w http.ResponseWriter, r *http.Request) {

	e.log.Println("Duplicate Match")
	http.ServeFile(w, r, "views/errors/duplicate.html") // make sure path is correct

}


func (e* ErrorHandler) UnAuthenticated(w http.ResponseWriter, r *http.Request) {

	e.log.Println("Unauthenticated")
	http.ServeFile(w, r, "views/errors/unauthenticated.html") // make sure path is correct

}

