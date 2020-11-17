package history

import (
	sq "github.com/Masterminds/squirrel"

	"github.com/hcnet/go/services/aurora/internal/db2"
	"github.com/hcnet/go/support/errors"
)

func (q *Q) GetAccountSignersByAccountID(id string) ([]AccountSigner, error) {
	sql := selectAccountSigners.
		Where(sq.Eq{"accounts_signers.account_id": id}).
		OrderBy("accounts_signers.signer asc")

	var results []AccountSigner
	if err := q.Select(&results, sql); err != nil {
		return nil, errors.Wrap(err, "could not run select query")
	}

	return results, nil
}

func (q *Q) SignersForAccounts(accounts []string) ([]AccountSigner, error) {
	sql := selectAccountSigners.
		Where(map[string]interface{}{"accounts_signers.account_id": accounts})

	var results []AccountSigner
	if err := q.Select(&results, sql); err != nil {
		return nil, errors.Wrap(err, "could not run select query")
	}

	return results, nil
}

// AccountsForSigner returns a list of `AccountSigner` rows for a given signer
func (q *Q) AccountsForSigner(signer string, page db2.PageQuery) ([]AccountSigner, error) {
	sql := selectAccountSigners.Where("accounts_signers.signer = ?", signer)
	sql, err := page.ApplyToUsingCursor(sql, "accounts_signers.account_id", page.Cursor)
	if err != nil {
		return nil, errors.Wrap(err, "could not apply query to page")
	}

	var results []AccountSigner
	if err := q.Select(&results, sql); err != nil {
		return nil, errors.Wrap(err, "could not run select query")
	}

	return results, nil
}

// CreateAccountSigner creates a row in the accounts_signers table.
// Returns number of rows affected and error.
func (q *Q) CreateAccountSigner(account, signer string, weight int32, sponsor *string) (int64, error) {
	sql := sq.Insert("accounts_signers").
		Columns("account_id", "signer", "weight", "sponsor").
		Values(account, signer, weight, sponsor)

	result, err := q.Exec(sql)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// RemoveAccountSigner deletes a row in the accounts_signers table.
// Returns number of rows affected and error.
func (q *Q) RemoveAccountSigner(account, signer string) (int64, error) {
	sql := sq.Delete("accounts_signers").Where(sq.Eq{
		"account_id": account,
		"signer":     signer,
	})

	result, err := q.Exec(sql)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

var selectAccountSigners = sq.Select("accounts_signers.*").From("accounts_signers")
