package api

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/middleware"
)

func Routes() {
	// router.HandleFunc("/createuser", CreateUser).Methods("POST") // TODO: integrate with auth0
	router := mux.NewRouter().StrictSlash(true)
	router.Use(middleware.EnsureValidToken())

	router.HandleFunc("/", homeLink)
	router.HandleFunc("/users", GetAllUsers).Methods("GET")
	router.HandleFunc("/balance/{userid}", GetUserBalanceFromHttp).Methods("GET")
	router.HandleFunc("/pendingWithdrawalTx/{userid}", GetPendingWithdrawalTx).Methods("GET")
	router.HandleFunc("/withdraw", Withdraw).Methods("POST")

	headers := handlers.AllowedHeaders([]string{
		"X-Requested-With",
		"Content-Type",
		"Authorization",
	})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{
		"http://localhost:3000",
		"https://zener-fe.vercel.app",
	})
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(router)))
}
