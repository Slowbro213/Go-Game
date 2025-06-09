package utils

import (
	"html/template"
	"net/http"
)

// MessageData holds all template variables
type MessageData struct {
	Type     string // "success", "error", "warning", "info"
	Title    string
	Message  string
	Link     string // optional
	LinkText string // optional
	AutoShow bool   // whether to auto-show the message via JavaScript
}

var templates *template.Template

func init() {
	// Parse templates once at startup
	templates = template.Must(template.ParseFiles("./views/message.html"))
}

func RenderMessage(w http.ResponseWriter, data MessageData) {
	w.Header().Set("Content-Type", "text/html")
	
	// Escape the message for JavaScript to prevent XSS
	escapedMessage := template.JSEscapeString(data.Message)
	data.Message = escapedMessage
	
	err := templates.ExecuteTemplate(w, "message.html", data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

// Example handler for success message
//func successHandler(w http.ResponseWriter, r *http.Request) {
//	renderMessage(w, MessageData{
//		Type:     "success",
//		Title:    "Operation Successful",
//		Message:  "Your changes have been saved successfully.",
//		Link:     "/dashboard",
//		LinkText: "Go to Dashboard",
//	})
//}
//
//// Example handler for error message
//func errorHandler(w http.ResponseWriter, r *http.Request) {
//	renderMessage(w, MessageData{
//		Type:     "error",
//		Title:    "Access Denied",
//		Message:  "You don't have permission to view this page.",
//		Link:     "/login",
//		LinkText: "Login Page",
//	})
//}
//
//// Generic handler that can show any message
//func messageHandler(w http.ResponseWriter, r *http.Request) {
//	// You could get these values from query params or session
//	msgType := r.URL.Query().Get("type")
//	if msgType == "" {
//		msgType = "info"
//	}
//
//	renderMessage(w, MessageData{
//		Type:     msgType,
//		Title:    r.URL.Query().Get("title"),
//		Message:  r.URL.Query().Get("message"),
//		Link:     r.URL.Query().Get("link"),
//		LinkText: r.URL.Query().Get("linkText"),
//	})
//}
//

