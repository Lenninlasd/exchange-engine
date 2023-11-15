package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/middleware"
)

func Routes() {
	router := mux.NewRouter().StrictSlash(true)
	supaRouter := router.PathPrefix("/supa").Subrouter()
	supaRouter.Use(middleware.ValidateTokenMiddleware)
	supaRouter.HandleFunc("/", homeLink).Methods("GET")
	supaRouter.HandleFunc("/signup", CreateUserAfterSignUpSupa).Methods("POST")
	supaRouter.HandleFunc("/balance", GetUserBalanceFromHttpSupa).Methods("GET")
	supaRouter.HandleFunc("/exchangerate/BTCCOP", ExchangerateApi).Methods("GET")
	supaRouter.HandleFunc("/mempool/transaction-fee", GetTransactionFee).Methods("GET")
	supaRouter.HandleFunc("/mempool/transaction-fee/{address}", GetTransactionFee).Methods("GET")
	supaRouter.HandleFunc("/reference", GetDepositReference).
		Queries("amount", "{amount:[0-9]+}", "currency", "{currency:[A-Za-z]+}").
		Methods("GET")
	supaRouter.HandleFunc("/deposit", CreateFiatDepositSupa).Methods("POST")
	supaRouter.HandleFunc("/pending-withdrawal-tx", GetPendingWithdrawalTxSupa).Methods("GET")
	supaRouter.HandleFunc("/withdraw", Withdraw).Methods("POST")

	privateRouter := router.PathPrefix("/prv").Subrouter()
	privateRouter.Use(middleware.EnsureValidToken())

	router.HandleFunc("/", homeLink).Methods("GET")
	router.HandleFunc("/wompi/event", UptateFiatDeposit).Methods("POST")
	router.HandleFunc("/balance/{userid}", GetUserBalanceFromHttp).Methods("GET")

	privateRouter.HandleFunc("/balance/{userid}", GetUserBalanceFromHttp).Methods("GET")
	privateRouter.HandleFunc("/pendingWithdrawalTx/{userid}", GetPendingWithdrawalTx).Methods("GET")
	privateRouter.HandleFunc("/reference", GetDepositReference).
		Queries("amount", "{amount:[0-9]+}", "currency", "{currency:[A-Za-z]+}").
		Methods("GET")
	privateRouter.HandleFunc("/mempool/transaction-fee", GetTransactionFee).Methods("GET")
	privateRouter.HandleFunc("/mempool/transaction-fee/{address}", GetTransactionFee).Methods("GET")

	privateRouter.HandleFunc("/exchangerate/BTCCOP", ExchangerateApi).Methods("GET")

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
		"https://zenerbtc.com",
	})
	var port = os.Getenv("ENGINE_PORT")
	if port == "" {
		port = ":8080"
	}
	log.Fatal(http.ListenAndServe(port, handlers.CORS(headers, methods, origins)(router)))
}
