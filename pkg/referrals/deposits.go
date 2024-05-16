package referrals

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lenninlasd/Engine/constants"
	"github.com/lenninlasd/Engine/postgres"
	"github.com/lenninlasd/Engine/types"
)

// * Check if the user who makes the deposit is valid to receive the referral bonus ...
// * or if the REFERRING user is valid to obtain the bonus.
func checkUserValidForReferral(db *sql.DB, userId string) (bool, error) {
	var userHasBeenReferred bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM public.referrals WHERE user_id=$1 AND referrer_id is not null)", userId).Scan(&userHasBeenReferred)
	if err != nil || !userHasBeenReferred {
		return false, err
	}

	var hasUsedReferralBonus bool
	query := `SELECT EXISTS(SELECT 1 FROM public.fiat_deposit_tx WHERE user_id=$1 AND origin=$2)`
	err = db.QueryRow(query, userId, constants.BY_REFERRAL_ORIGIN).Scan(&hasUsedReferralBonus)
	if err != nil || hasUsedReferralBonus {
		return false, err
	}

	return true, nil
}

func CreateDepositToReferredUser(userId string, transaction types.WompiTx, btcCurrentPriceInCents uint64) {
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

	result, err := tx.Exec(`UPDATE btcbalances SET amount=amount+$1 WHERE user_id=$2;`, newPurchasedSats, userId)
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
		panic(errors.New("CreateDepositToReferredUser wrong event: No rows in btc balance affected"))
	}

	tx.Commit()
}

func getReferringUser(db *sql.DB, referredUserId string) (string, error) {
	var userId string
	err := db.QueryRow(`
		SELECT users.id as referrer_user_id
		FROM public.referrals
		JOIN public.users ON users.referral_id = referrer_id
		WHERE user_id=$1`, referredUserId).Scan(&userId)

	return userId, err
}

func CreateDepositForReferring(referredUserId string, transaction types.WompiTx, btcCurrentPriceInCents uint64) {
	if transaction.AmountInCents < constants.MIN_REFFERAL_AMOUNT {
		return
	}

	db := postgres.OpenConnection()
	defer db.Close()

	isValidForReferral, err := checkUserValidForReferral(db, referredUserId)
	if err != nil {
		panic(err)
	}
	if !isValidForReferral {
		return
	}

	userId, err := getReferringUser(db, referredUserId)
	if err != nil {
		panic(err)
	}

	reference := uuid.New().String()
	newAmountInCents := constants.MIN_REFFERAL_AMOUNT / 10
	newPurchasedSats := constants.SATS_PER_BTC * newAmountInCents / btcCurrentPriceInCents

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	// * the new `id` is the reference of the main deposit_tx
	_, err = tx.Exec(
		`INSERT INTO fiat_deposit_tx (reference, id, amount_in_cents, purchased_sats, status, user_id, origin)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, reference, transaction.Reference, newAmountInCents,
		newPurchasedSats, constants.APPROVED, userId, constants.REFERRING_ORIGIN,
	)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	result, err := tx.Exec(`UPDATE btcbalances SET amount=amount+$1 WHERE user_id=$2;`, newPurchasedSats, userId)
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
		panic(errors.New("CreateDepositToReferredUser wrong event: No rows in btc balance affected"))
	}

	tx.Commit()
}
