package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	"github.com/victorprb/gochat/pkg/trace"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse()

	// setup goth
	gothic.GetProviderName = func(r *http.Request) (string, error) {
		provider := strings.Split(r.URL.Path, "/")[2]

		return provider, nil
	}

	goth.UseProviders(
		google.New(os.Getenv("GOCHAT_GOOGLE_KEY"), os.Getenv("GOCHAT_GOOGLE_SECRET"),
			"http://localhost:8080/auth/google/callback"),
	)

	r := newRoom(UseFileSystemAvatar)
	r.tracer = trace.New(os.Stdout)

	// routes
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/upload", &templateHandler{filename: "upload.html"})
	http.HandleFunc("/uploader", uploaderHandler)
	http.Handle("/avatars/",
		http.StripPrefix("/avatars/",
			http.FileServer(http.Dir("./avatars"))))

	http.Handle("/room", r)

	// creates a room
	go r.run()

	// start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("./web/template", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}

	if authCookie, err := r.Cookie("auth"); err == nil {
		userJson, err := base64.StdEncoding.DecodeString(authCookie.Value)
		if err != nil {
			fmt.Println("decode error: ", err)
			return
		}

		var user UserData
		err = json.Unmarshal(userJson, &user)
		if err != nil {
			log.Fatal("Failed to json decode: ", err)
			return
		}

		data["UserData"] = user
	}

	t.templ.Execute(w, data)
}
