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
	privateRouter := router.PathPrefix("/prv").Subrouter()
	privateRouter.Use(middleware.EnsureValidToken())

	router.HandleFunc("/", homeLink).Methods("GET")
	router.HandleFunc("/wompi/event", UptateFiatDeposit).Methods("POST")

	privateRouter.HandleFunc("/users", GetAllUsers).Methods("GET") // TODO: only for admin.
	privateRouter.HandleFunc("/balance/{userid}", GetUserBalanceFromHttp).Methods("GET")
	privateRouter.HandleFunc("/pendingWithdrawalTx/{userid}", GetPendingWithdrawalTx).Methods("GET")
	privateRouter.HandleFunc("/reference", GetDepositReference).
		Queries("amount", "{amount:[0-9]+}", "currency", "{currency:[A-Za-z]+}").
		Methods("GET")
	privateRouter.HandleFunc("/mempool/transaction-fee", GetTransactionFee).Methods("GET")
	privateRouter.HandleFunc("/mempool/transaction-fee/{address}", GetTransactionFee).Methods("GET")

	privateRouter.HandleFunc("/exchangerate/BTCCOP", Exchangerate).Methods("GET")

	privateRouter.HandleFunc("/withdraw", Withdraw).Methods("POST")
	privateRouter.HandleFunc("/signup", CreateUserAfterSignUp).Methods("POST")
	privateRouter.HandleFunc("/deposit", CreateFiatDeposit).Methods("POST")

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
