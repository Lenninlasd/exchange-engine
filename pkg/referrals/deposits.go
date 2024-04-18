package referrals

import (
	"errors"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

const REFERRAL_ORIGIN = "REFERRAL"
const APPROVED = "APPROVED"
const MIN_REFFERAL_AMOUNT uint64 = 10000000
const SATS_PER_BTC uint64 = 100000000

func CreateDepositReferral(userId string, transaction types.WompiTx, btcCurrentPriceInCents uint64, balanceAmountInSats uint64) {
	// TODO: this user has referrals?
	// check if was refered
	// check if do not have first referal

	if transaction.AmountInCents < MIN_REFFERAL_AMOUNT {
		return
	}

	db := postgres.OpenConnection()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	reference := uuid.New().String()
	newAmountInCents := MIN_REFFERAL_AMOUNT / 10
	newPurchasedSats := SATS_PER_BTC * newAmountInCents / btcCurrentPriceInCents

	// * the new `id` is the reference of the main deposit_tx
	_, err = tx.Exec(
		`INSERT INTO fiat_deposit_tx (reference, id, amount_in_cents, purchased_sats, status, user_id, origin)
		VALUES ($1, $2, $3, $4, $5, $6, 'REFERRAL')`, reference, transaction.Reference, newAmountInCents,
		newPurchasedSats, APPROVED, userId,
	)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	newBalance := newPurchasedSats + balanceAmountInSats
	result, err := tx.Exec(`UPDATE btcbalances SET amount=$1 WHERE user_id=$2;`, newBalance, userId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	numRowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	if numRowsAffected != 1 {
		tx.Rollback()
		panic(errors.New("UpdateBalanceAfterDeposit wrong event: No rows in btc balance affected"))
	}

	tx.Commit()
}
