package resourceadapter

import (
	"github.com/hcnet/go/amount"
	protocol "github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/assets"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

func PopulateBalance(dest *protocol.Balance, row history.TrustLine) (err error) {
	dest.Type, err = assets.String(row.AssetType)
	if err != nil {
		return errors.Wrap(err, "getting the string representation from the provided xdr asset type")
	}

	dest.Balance = amount.StringFromInt64(row.Balance)
	dest.BuyingLiabilities = amount.StringFromInt64(row.BuyingLiabilities)
	dest.SellingLiabilities = amount.StringFromInt64(row.SellingLiabilities)
	dest.Limit = amount.StringFromInt64(row.Limit)
	dest.Issuer = row.AssetIssuer
	dest.Code = row.AssetCode
	dest.LastModifiedLedger = row.LastModifiedLedger
	isAuthorized := row.IsAuthorized()
	dest.IsAuthorized = &isAuthorized
	dest.IsAuthorizedToMaintainLiabilities = &isAuthorized
	isAuthorizedToMaintainLiabilities := row.IsAuthorizedToMaintainLiabilities()
	if isAuthorizedToMaintainLiabilities {
		dest.IsAuthorizedToMaintainLiabilities = &isAuthorizedToMaintainLiabilities
	}
	if row.Sponsor.Valid {
		dest.Sponsor = row.Sponsor.String
	}
	return
}

func PopulateNativeBalance(dest *protocol.Balance, stroops, buyingLiabilities, sellingLiabilities xdr.Int64) (err error) {
	dest.Type, err = assets.String(xdr.AssetTypeAssetTypeNative)
	if err != nil {
		return errors.Wrap(err, "getting the string representation from the provided xdr asset type")
	}

	dest.Balance = amount.String(stroops)
	dest.BuyingLiabilities = amount.String(buyingLiabilities)
	dest.SellingLiabilities = amount.String(sellingLiabilities)
	dest.LastModifiedLedger = 0
	dest.Limit = ""
	dest.Issuer = ""
	dest.Code = ""
	dest.IsAuthorized = nil
	dest.IsAuthorizedToMaintainLiabilities = nil
	return
}
