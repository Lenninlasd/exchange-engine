package constants

import "os"

const (
	ZENER_FEE uint64 = 900

	WITHDRAW_STATUS_INICIALIZED = "initialized"
	WITHDRAW_STATUS_BROADCASTED = "broadcasted"
	WITHDRAW_STATUS_CONFIRMED   = "confirmed"
	WITHDRAW_STATUS_CANCELED    = "canceled"
	WITHDRAW_INSUFFICIENT_FUNDS = "InsufficientFunds"

	TIME_TO_FECHT_MEMPOOL = 1

	ADMIN_ROLE = "admin"
)

var WALLET_ADDRESS = os.Getenv("WALLET_ADDRESS")
var BITCOIN_NETWORK = os.Getenv("BITCOIN_NETWORK")
