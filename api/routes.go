package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lenninlasd/Engine/api/admin"
	"github.com/lenninlasd/Engine/api/deposit"
	"github.com/lenninlasd/Engine/api/withdrawal"
	"github.com/lenninlasd/Engine/middleware"
)

// Routes returns a configured router
func Routes(validateMiddleware func(http.Handler) http.Handler) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	supaRouter := router.PathPrefix("/supa").Subrouter()
	supaRouter.Use(validateMiddleware) // Injected middleware
	supaRouter.HandleFunc("/", homeLink).Methods("GET")
	supaRouter.HandleFunc("/signup", CreateUserAfterSignUpSupa).Methods("POST")
	supaRouter.HandleFunc("/balance", GetUserBalanceFromHttpSupa).Methods("GET")
	supaRouter.HandleFunc("/user", GetUser).Methods("GET")

	supaRouter.HandleFunc("/exchangerate/BTCCOP", ExchangeRateApi).Methods("GET")
	supaRouter.HandleFunc("/mempool/transaction-fee", GetTransactionFee).Methods("GET")
	supaRouter.HandleFunc("/mempool/transaction-fee/{address}", GetTransactionFee).Methods("GET")
	supaRouter.HandleFunc("/reference", GetDepositReference).
		Queries("amount", "{amount:[0-9]+}", "currency", "{currency:[A-Za-z]+}").
		Methods("GET")

	// Deposit
	supaRouter.HandleFunc("/deposit-history", GetDepositHistory).Methods("GET")
	supaRouter.HandleFunc("/approved-deposit-history", admin.GetApprovedDepositHistory).Methods("GET")
	supaRouter.HandleFunc("/deposit", CreateFiatDepositSupa).Methods("POST")
	supaRouter.HandleFunc("/fund-backup", admin.UpdateFundBackup).Methods("PATCH")
	supaRouter.HandleFunc("/update-expired-deposits", deposit.UpdatePendingExpiredDeposits).Methods("PATCH")

	// Withdrawals
	supaRouter.HandleFunc("/pending-withdrawal-tx", GetPendingWithdrawalTxSupa).Methods("GET")
	supaRouter.HandleFunc("/withdraw", Withdraw).Methods("POST")
	supaRouter.HandleFunc("/withdrawals-history", withdrawal.GetWithdrawalsHistory).Methods("GET")
	supaRouter.HandleFunc("/blocked-withdrawals", withdrawal.GetBlockedWithdrawals).Methods("GET")
	supaRouter.HandleFunc("/resume-withdrawal/{withdrawalId}", withdrawal.ResumeWithdrawal).Methods("POST")

	// Referrals
	supaRouter.HandleFunc("/referral-history", GetReferralHistory).Methods("GET")

	router.HandleFunc("/", homeLink).Methods("GET")
	router.HandleFunc("/referrer/{id}", CheckReferrerUser).Methods("GET")
	router.HandleFunc("/wompi/event", UptateFiatDeposit).Methods("POST")

	headers := handlers.AllowedHeaders([]string{
		"X-Requested-With",
		"Content-Type",
		"Authorization",
	})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{
		"http://localhost:3000",
		"https://zener-fe.vercel.app",
		"https://zenerbtc.com",
	})

	// Apply CORS middleware
	return handlers.CORS(headers, methods, origins)(router)
}

// StartServer runs the HTTP server
func StartServer() {
	router := Routes(middleware.ValidateTokenMiddleware)
	port := os.Getenv("ENGINE_PORT")
	if port == "" {
		port = ":8080"
	}
	log.Fatal(http.ListenAndServe(port, router))
}
