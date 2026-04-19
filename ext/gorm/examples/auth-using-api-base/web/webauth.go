package web

import (
	"fmt"

	"github.com/pageton/gotg"
)

type webAuth struct {
	phoneChan     chan string
	codeChan      chan string
	passwdChan    chan string
	phoneWritten  chan struct{}
	codeWritten   chan struct{}
	passwdWritten chan struct{}
	authStatus    gotg.AuthStatus
}

func GetWebAuth() *webAuth {
	return &webAuth{
		phoneChan:     make(chan string),
		codeChan:      make(chan string),
		passwdChan:    make(chan string),
		phoneWritten:  make(chan struct{}, 1),
		codeWritten:   make(chan struct{}, 1),
		passwdWritten: make(chan struct{}, 1),
	}
}

func (w *webAuth) AskPhoneNumber() (string, error) {
	if w.authStatus.Event == gotg.AuthStatusPhoneRetrial {
		fmt.Println("The phone number you just entered seems to be incorrect,")
		fmt.Println("Attempts Left:", w.authStatus.AttemptsLeft)
		fmt.Println("Please try again....")
	}
	fmt.Println("waiting for phone...")
	code := <-w.phoneChan
	return code, nil
}

func (w *webAuth) AskCode() (string, error) {
	if w.authStatus.Event == gotg.AuthStatusPhoneCodeRetrial {
		fmt.Println("The OTP you just entered seems to be incorrect,")
		fmt.Println("Attempts Left:", w.authStatus.AttemptsLeft)
		fmt.Println("Please try again....")
	}
	fmt.Println("waiting for code...")
	code := <-w.codeChan
	return code, nil
}

func (w *webAuth) AskPassword() (string, error) {
	if w.authStatus.Event == gotg.AuthStatusPasswordRetrial {
		fmt.Println("The 2FA password you just entered seems to be incorrect,")
		fmt.Println("Attempts Left:", w.authStatus.AttemptsLeft)
		fmt.Println("Please try again....")
	}
	fmt.Println("waiting for 2fa password...")
	code := <-w.passwdChan
	return code, nil
}

func (w *webAuth) AuthStatus(authStatusIp gotg.AuthStatus) {
	w.authStatus = authStatusIp
	// Signal the HTTP handler that the auth status has been updated.
	switch authStatusIp.Event {
	case gotg.AuthStatusPhoneAsked, gotg.AuthStatusPhoneRetrial, gotg.AuthStatusPhoneFailed,
		gotg.AuthStatusPhoneCodeAsked, gotg.AuthStatusPhoneCodeVerified, gotg.AuthStatusPhoneCodeRetrial, gotg.AuthStatusPhoneCodeFailed,
		gotg.AuthStatusPasswordAsked, gotg.AuthStatusPasswordRetrial, gotg.AuthStatusPasswordFailed,
		gotg.AuthStatusSuccess:
		// Drain previous signal if any, then send new one
		select {
		case <-w.phoneWritten:
		default:
		}
		select {
		case <-w.codeWritten:
		default:
		}
		select {
		case <-w.passwdWritten:
		default:
		}
		w.phoneWritten <- struct{}{}
		w.codeWritten <- struct{}{}
		w.passwdWritten <- struct{}{}
	}
}

func (w *webAuth) ReceivePhone(phone string) {
	w.phoneChan <- phone
}

func (w *webAuth) ReceiveCode(code string) {
	w.codeChan <- code
}

func (w *webAuth) ReceivePasswd(passwd string) {
	w.passwdChan <- passwd
}
