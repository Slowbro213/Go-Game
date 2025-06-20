package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"


	"game/utils"

	"github.com/gorilla/sessions"
)

type SessionHandler struct {
	store *sessions.CookieStore
	log   *log.Logger
}

const (
	cookieName    = "poked-cookie"
	sessionMaxAge = 3600 * 2 // 2 hours in seconds
)

func NewSessionHandler(s *sessions.CookieStore, l *log.Logger) *SessionHandler {
	return &SessionHandler{store: s, log: l}
}

func (s *SessionHandler) getSession(r *http.Request) (*sessions.Session, error) {
	session, err := s.store.Get(r, cookieName)
	if err != nil {
		s.log.Printf("Session retrieval error: %v", err)
		return nil, fmt.Errorf("session error")
	}
	return session, nil
}

func (s *SessionHandler) Secret(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSession(r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check both authentication and session freshness
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		s.log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Add session activity tracking
	session.Values["last_activity"] = time.Now().Unix()
	if err := session.Save(r, w); err != nil {
		s.log.Printf("Session save error: %v", err)
	}

	s.log.Printf("Secret accessed by %s", r.RemoteAddr)
	fmt.Fprintln(w, "The cake is a lie!")
}

func (s *SessionHandler) Login(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSession(r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Regenerate session ID on login to prevent session fixation
	session.Options.MaxAge = sessionMaxAge
	session.Values["authenticated"] = true
	session.Values["player_id"] = utils.GenerateToken()
	session.Values["login_time"] = time.Now().Unix()
	session.Values["user_agent"] = r.UserAgent()
	session.Values["ip_address"] = r.RemoteAddr

	if err := session.Save(r, w); err != nil {
		s.log.Printf("Login session save error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.Printf("Successful login from %s", r.RemoteAddr)
	utils.RenderMessage(w, utils.MessageData{
		Type:     "success",
		Title:    "Login Successful",
		Message:  "You are now logged in",
		Link:     "/",
		LinkText: "Go to Game",

	})
}


func (s *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSession(r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Clear all session values and set MaxAge to -1 to delete cookie
	for k := range session.Values {
		delete(session.Values, k)
	}
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		s.log.Printf("Logout session save error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.Printf("Successful logout from %s", r.RemoteAddr)
	utils.RenderMessage(w, utils.MessageData{
			Type:     "success",
			Title:    "Logout Successful",
			Message:  "You have Logged Out",
			Link:     "/login",
			LinkText: "Login",
	
		})
}

// Add this middleware to check session validity on protected routes
func (s *SessionHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := s.getSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check authentication
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			s.log.Printf("Unauthenticated access attempt to %s from %s", r.URL.Path, r.RemoteAddr)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check session freshness (optional)
		if loginTime, ok := session.Values["login_time"].(int64); ok {
			if time.Now().Unix()-loginTime > sessionMaxAge {
				s.log.Printf("Expired session attempt from %s", r.RemoteAddr)
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	}
}
