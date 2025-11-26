package main

import (
	"ThisBot/common"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func poll_handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func logout_handler(w http.ResponseWriter, r *http.Request) {

}

func login_handler(w http.ResponseWriter, r *http.Request) {

}

func Server() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/poll", poll_handler)
	router.Post("/login", login_handler)
	router.Post("/logout", logout_handler)

	strPort := strconv.Itoa(common.Cfg.Server.Port)
	log.Println("[+] Server running on " + common.Cfg.Server.Host + ":" + strPort)

	http.ListenAndServe(":"+strPort, router)
}
