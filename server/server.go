package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func poll_handler(w http.ResponseWriter, r *http.Request) {

}

func logout_handler(w http.ResponseWriter, r *http.Request) {

}

func login_handler(w http.ResponseWriter, r *http.Request) {

}

func Server() {
	router := chi.NewRouter()

	router.HandleFunc("poll", poll_handler)
	router.HandleFunc("login", login_handler)
	router.HandleFunc("logout", logout_handler)
}
