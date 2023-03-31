package api

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func Routes() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	// router.HandleFunc("/createuser", CreateUser).Methods("POST") // TODO: integrate with auth0
	router.HandleFunc("/users", GetAllUsers).Methods("GET")
	// router.HandleFunc("/countusers", CountAllUsers).Methods("GET")
	// router.HandleFunc("/user/{id}", GetOneUser).Methods("GET")
	// router.HandleFunc("/removeuser/{id}", DeleteOneUser).Methods("DELETE")
	// router.HandleFunc("/updateuser/{id}", UpdateUser).Methods("PATCH")

	router.HandleFunc("/balance/{userid}", GetUserBalanceFromHttp).Methods("GET")
	// router.HandleFunc("/createbalance", CreateBalance).Methods("POST")

	// router.HandleFunc("/orderbook", GetOrderBook).Methods("GET")
	// router.HandleFunc("/createOrder", CreateOrder).Methods("POST")

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
