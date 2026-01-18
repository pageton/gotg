package web

import (
	"fmt"
	"net/http"

	"github.com/pageton/gotg"
)

// Start a web server and wait
func Start(wa *webAuth) {
	http.HandleFunc("/", wa.setInfo)
	http.HandleFunc("/getAuthStatus", wa.getAuthStatus)
	http.ListenAndServe(":9997", nil)
}

func (wa *webAuth) getAuthStatus(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, wa.authStatus.Event)
}

// setInfo handle user info, set phone, code or passwd
func (wa *webAuth) setInfo(w http.ResponseWriter, req *http.Request) {
	action := req.URL.Query().Get("set")

	switch action {

	case "phone":
		fmt.Println("Rec phone")
		num := req.URL.Query().Get("phone")
		phone := "+" + num
		wa.ReceivePhone(phone)
		for wa.authStatus.Event == gotg.AuthStatusPhoneAsked ||
			wa.authStatus.Event == gotg.AuthStatusPhoneRetrial {
			continue
		}
	case "code":
		fmt.Println("Rec code")
		code := req.URL.Query().Get("code")
		wa.ReceiveCode(code)
		for wa.authStatus.Event == gotg.AuthStatusPhoneCodeAsked ||
			wa.authStatus.Event == gotg.AuthStatusPhoneCodeRetrial {
			continue
		}
	case "passwd":
		passwd := req.URL.Query().Get("passwd")
		wa.ReceivePasswd(passwd)
		for wa.authStatus.Event == gotg.AuthStatusPasswordAsked ||
			wa.authStatus.Event == gotg.AuthStatusPasswordRetrial {
			continue
		}
	}
	w.Write([]byte(wa.authStatus.Event))
}
