package api

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/middleware"
)

func Routes() {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(middleware.EnsureValidToken())

	router.HandleFunc("/", homeLink)
	router.HandleFunc("/users", GetAllUsers).Methods("GET") // TODO: only for admin.
	router.HandleFunc("/balance/{userid}", GetUserBalanceFromHttp).Methods("GET")
	router.HandleFunc("/pendingWithdrawalTx/{userid}", GetPendingWithdrawalTx).Methods("GET")
	router.HandleFunc("/reference", GetDepositReference).
		Queries("amount", "{amount:[0-9]+}", "currency", "{currency:[A-Za-z]+}").
		Methods("GET")

	router.HandleFunc("/withdraw", Withdraw).Methods("POST")
	router.HandleFunc("/signup", CreateUserAfterSignUp).Methods("POST")
	router.HandleFunc("/deposit", CreateFiatDeposit).Methods("POST")

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
