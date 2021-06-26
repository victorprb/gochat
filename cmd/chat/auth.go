package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/markbates/goth/gothic"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")
	if err == http.ErrNoCookie || cookie.Value == "" {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.next.ServeHTTP(w, r)
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "callback") {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		userData := UserData{"name": user.Name, "avatar_url": user.AvatarURL, "email": user.Email}
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: base64EncodeUserData(userData),
			Path:  "/"})

		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	// try to get the user without re-authenticating
	if user, err := gothic.CompleteUserAuth(w, r); err == nil {
		userData := UserData{"name": user.Name, "avatar_url": user.AvatarURL, "email": user.Email}
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: base64EncodeUserData(userData),
			Path:  "/"})

		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func base64EncodeUserData(u UserData) string {
	userJson, err := json.Marshal(u)
	if err != nil {
		log.Fatal("Could not json encode: ", err)
	}

	return base64.StdEncoding.EncodeToString(userJson)
}
