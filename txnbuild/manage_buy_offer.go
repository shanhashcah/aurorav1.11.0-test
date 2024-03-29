package txnbuild

import (
	"github.com/hcnet/go/amount"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

// ManageBuyOffer represents the Hcnet manage buy offer operation. See
// https://www.hcnet.org/developers/guides/concepts/list-of-operations.html
type ManageBuyOffer struct {
	Selling       Asset
	Buying        Asset
	Amount        string
	Price         string
	price         price
	OfferID       int64
	SourceAccount Account
}

// BuildXDR for ManageBuyOffer returns a fully configured XDR Operation.
func (mo *ManageBuyOffer) BuildXDR() (xdr.Operation, error) {
	xdrSelling, err := mo.Selling.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set XDR 'Selling' field")
	}

	xdrBuying, err := mo.Buying.ToXDR()
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to set XDR 'Buying' field")
	}

	xdrAmount, err := amount.Parse(mo.Amount)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse 'Amount'")
	}

	if err = mo.price.parse(mo.Price); err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to parse 'Price'")
	}

	opType := xdr.OperationTypeManageBuyOffer
	xdrOp := xdr.ManageBuyOfferOp{
		Selling:   xdrSelling,
		Buying:    xdrBuying,
		BuyAmount: xdrAmount,
		Price:     mo.price.toXDR(),
		OfferId:   xdr.Int64(mo.OfferID),
	}
	body, err := xdr.NewOperationBody(opType, xdrOp)
	if err != nil {
		return xdr.Operation{}, errors.Wrap(err, "failed to build XDR OperationBody")
	}

	op := xdr.Operation{Body: body}
	SetOpSourceAccount(&op, mo.SourceAccount)
	return op, nil
}

// FromXDR for ManageBuyOffer initialises the txnbuild struct from the corresponding xdr Operation.
func (mo *ManageBuyOffer) FromXDR(xdrOp xdr.Operation) error {
	result, ok := xdrOp.Body.GetManageBuyOfferOp()
	if !ok {
		return errors.New("error parsing manage_buy_offer operation from xdr")
	}

	mo.SourceAccount = accountFromXDR(xdrOp.SourceAccount)
	mo.OfferID = int64(result.OfferId)
	mo.Amount = amount.String(result.BuyAmount)
	if result.Price != (xdr.Price{}) {
		mo.price.fromXDR(result.Price)
		mo.Price = mo.price.string()
	}
	buyingAsset, err := assetFromXDR(result.Buying)
	if err != nil {
		return errors.Wrap(err, "error parsing buying_asset in manage_buy_offer operation")
	}
	mo.Buying = buyingAsset

	sellingAsset, err := assetFromXDR(result.Selling)
	if err != nil {
		return errors.Wrap(err, "error parsing selling_asset in manage_buy_offer operation")
	}
	mo.Selling = sellingAsset
	return nil
}

// Validate for ManageBuyOffer validates the required struct fields. It returns an error if any
// of the fields are invalid. Otherwise, it returns nil.
func (mo *ManageBuyOffer) Validate() error {
	return validateOffer(mo.Buying, mo.Selling, mo.Amount, mo.Price, mo.OfferID)
}

// GetSourceAccount returns the source account of the operation, or nil if not
// set.
func (mo *ManageBuyOffer) GetSourceAccount() Account {
	return mo.SourceAccount
}
