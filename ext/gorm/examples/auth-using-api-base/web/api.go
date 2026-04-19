package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SECURITY WARNING: This example demonstrates the auth flow but is NOT production-ready.
// Before deploying: add TLS, authentication middleware, CSRF protection, and rate limiting.
// Never expose this on a public network without these safeguards.

type authRequest struct {
	Phone string `json:"phone,omitempty"`
	Code  string `json:"code,omitempty"`
	Pass  string `json:"pass,omitempty"`
}

// Start a web server and wait.
// Binds to localhost only (127.0.0.1) to limit network exposure.
func Start(wa *webAuth) {
	mux := http.NewServeMux()
	mux.HandleFunc("/getAuthStatus", wa.getAuthStatus)
	mux.HandleFunc("/", wa.setInfo)

	server := &http.Server{
		Addr:              "127.0.0.1:9997",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Println("Web auth API listening on http://127.0.0.1:9997")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("Web server error:", err)
	}
}

func (wa *webAuth) getAuthStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprint(w, wa.authStatus.Event)
}

func (wa *webAuth) setInfo(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body authRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	switch {
	case body.Phone != "":
		phone := "+" + body.Phone
		wa.ReceivePhone(phone)
		<-wa.phoneWritten
	case body.Code != "":
		wa.ReceiveCode(body.Code)
		<-wa.codeWritten
	case body.Pass != "":
		wa.ReceivePasswd(body.Pass)
		<-wa.passwdWritten
	default:
		http.Error(w, "provide one of: phone, code, pass", http.StatusBadRequest)
		return
	}

	w.Write([]byte(wa.authStatus.Event))
}
