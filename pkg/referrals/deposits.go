package referrals

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

func checkUserValidForReferral(db *sql.DB, userId string) (bool, error) {
	var userHasBeenReferred bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM public.referrals WHERE user_id=$1 AND referrer_id is not null)", userId).Scan(&userHasBeenReferred)
	if err != nil || !userHasBeenReferred {
		return false, err
	}

	var hasUsedReferralBonus bool
	query := `SELECT EXISTS(SELECT 1 FROM public.fiat_deposit_tx WHERE user_id=$1 AND origin=$2))`
	err = db.QueryRow(query, userId, constants.BY_REFERRAL_ORIGIN).Scan(&hasUsedReferralBonus)
	if err != nil || hasUsedReferralBonus {
		return false, err
	}

	return true, nil
}

func CreateDepositReferral(userId string, transaction types.WompiTx, btcCurrentPriceInCents uint64, balanceAmountInSats uint64) {
	if transaction.AmountInCents < constants.MIN_REFFERAL_AMOUNT {
		return
	}

	db := postgres.OpenConnection()
	defer db.Close()

	isValidForReferral, err := checkUserValidForReferral(db, userId)
	if err != nil {
		panic(err)
	}
	if !isValidForReferral {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	reference := uuid.New().String()
	newAmountInCents := constants.MIN_REFFERAL_AMOUNT / 10
	newPurchasedSats := constants.SATS_PER_BTC * newAmountInCents / btcCurrentPriceInCents

	// * the new `id` is the reference of the main deposit_tx
	_, err = tx.Exec(
		`INSERT INTO fiat_deposit_tx (reference, id, amount_in_cents, purchased_sats, status, user_id, origin)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, reference, transaction.Reference, newAmountInCents,
		newPurchasedSats, constants.APPROVED, userId, constants.BY_REFERRAL_ORIGIN,
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
