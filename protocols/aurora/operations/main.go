package operations

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/protocols/aurora/base"
	"github.com/hcnet/go/support/render/hal"
	"github.com/hcnet/go/xdr"
)

// TypeNames maps from operation type to the string used to represent that type
// in aurora's JSON responses
var TypeNames = map[xdr.OperationType]string{
	xdr.OperationTypeCreateAccount:                 "create_account",
	xdr.OperationTypePayment:                       "payment",
	xdr.OperationTypePathPaymentStrictReceive:      "path_payment_strict_receive",
	xdr.OperationTypeManageSellOffer:               "manage_sell_offer",
	xdr.OperationTypeCreatePassiveSellOffer:        "create_passive_sell_offer",
	xdr.OperationTypeSetOptions:                    "set_options",
	xdr.OperationTypeChangeTrust:                   "change_trust",
	xdr.OperationTypeAllowTrust:                    "allow_trust",
	xdr.OperationTypeAccountMerge:                  "account_merge",
	xdr.OperationTypeInflation:                     "inflation",
	xdr.OperationTypeManageData:                    "manage_data",
	xdr.OperationTypeBumpSequence:                  "bump_sequence",
	xdr.OperationTypeManageBuyOffer:                "manage_buy_offer",
	xdr.OperationTypePathPaymentStrictSend:         "path_payment_strict_send",
	xdr.OperationTypeCreateClaimableBalance:        "create_claimable_balance",
	xdr.OperationTypeClaimClaimableBalance:         "claim_claimable_balance",
	xdr.OperationTypeBeginSponsoringFutureReserves: "begin_sponsoring_future_reserves",
	xdr.OperationTypeEndSponsoringFutureReserves:   "end_sponsoring_future_reserves",
	xdr.OperationTypeRevokeSponsorship:             "revoke_sponsorship",
}

// Base represents the common attributes of an operation resource
type Base struct {
	Links struct {
		Self        hal.Link `json:"self"`
		Transaction hal.Link `json:"transaction"`
		Effects     hal.Link `json:"effects"`
		Succeeds    hal.Link `json:"succeeds"`
		Precedes    hal.Link `json:"precedes"`
	} `json:"_links"`

	ID string `json:"id"`
	PT string `json:"paging_token"`
	// TransactionSuccessful defines if this operation is part of
	// successful transaction.
	TransactionSuccessful bool      `json:"transaction_successful"`
	SourceAccount         string    `json:"source_account"`
	Type                  string    `json:"type"`
	TypeI                 int32     `json:"type_i"`
	LedgerCloseTime       time.Time `json:"created_at"`
	// TransactionHash is the hash of the transaction which created the operation
	// Note that the Transaction field below is not always present in the Operation response.
	// If the Transaction field is present TransactionHash is redundant since the same information
	// is present in Transaction. But, if the Transaction field is nil then TransactionHash is useful.
	// Transaction is non nil when the "join=transactions" parameter is present in the operations request
	TransactionHash string               `json:"transaction_hash"`
	Transaction     *aurora.Transaction `json:"transaction,omitempty"`
	Sponsor         string               `json:"sponsor,omitempty"`
}

// PagingToken implements hal.Pageable
func (base Base) PagingToken() string {
	return base.PT
}

// BumpSequence is the json resource representing a single operation whose type is
// BumpSequence.
type BumpSequence struct {
	Base
	BumpTo string `json:"bump_to"`
}

// CreateAccount is the json resource representing a single operation whose type
// is CreateAccount.
type CreateAccount struct {
	Base
	StartingBalance string `json:"starting_balance"`
	Funder          string `json:"funder"`
	Account         string `json:"account"`
}

// Payment is the json resource representing a single operation whose type is
// Payment.
type Payment struct {
	Base
	base.Asset
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

// PathPayment is the json resource representing a single operation whose type
// is PathPayment.
type PathPayment struct {
	Payment
	Path              []base.Asset `json:"path"`
	SourceAmount      string       `json:"source_amount"`
	SourceMax         string       `json:"source_max"`
	SourceAssetType   string       `json:"source_asset_type"`
	SourceAssetCode   string       `json:"source_asset_code,omitempty"`
	SourceAssetIssuer string       `json:"source_asset_issuer,omitempty"`
}

// PathPaymentStrictSend is the json resource representing a single operation whose type
// is PathPaymentStrictSend.
type PathPaymentStrictSend struct {
	Payment
	Path              []base.Asset `json:"path"`
	SourceAmount      string       `json:"source_amount"`
	DestinationMin    string       `json:"destination_min"`
	SourceAssetType   string       `json:"source_asset_type"`
	SourceAssetCode   string       `json:"source_asset_code,omitempty"`
	SourceAssetIssuer string       `json:"source_asset_issuer,omitempty"`
}

// ManageData represents a ManageData operation as it is serialized into json
// for the aurora API.
type ManageData struct {
	Base
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Offer is an embedded resource used in offer type operations.
type Offer struct {
	Base
	Amount             string     `json:"amount"`
	Price              string     `json:"price"`
	PriceR             base.Price `json:"price_r"`
	BuyingAssetType    string     `json:"buying_asset_type"`
	BuyingAssetCode    string     `json:"buying_asset_code,omitempty"`
	BuyingAssetIssuer  string     `json:"buying_asset_issuer,omitempty"`
	SellingAssetType   string     `json:"selling_asset_type"`
	SellingAssetCode   string     `json:"selling_asset_code,omitempty"`
	SellingAssetIssuer string     `json:"selling_asset_issuer,omitempty"`
}

// CreatePassiveSellOffer is the json resource representing a single operation whose
// type is CreatePassiveSellOffer.
type CreatePassiveSellOffer struct {
	Offer
}

// ManageSellOffer is the json resource representing a single operation whose type
// is ManageSellOffer.
type ManageSellOffer struct {
	Offer
	OfferID int64 `json:"offer_id,string"`
}

// ManageBuyOffer is the json resource representing a single operation whose type
// is ManageBuyOffer.
type ManageBuyOffer struct {
	Offer
	OfferID int64 `json:"offer_id,string"`
}

// SetOptions is the json resource representing a single operation whose type is
// SetOptions.
type SetOptions struct {
	Base
	HomeDomain    string `json:"home_domain,omitempty"`
	InflationDest string `json:"inflation_dest,omitempty"`

	MasterKeyWeight *int   `json:"master_key_weight,omitempty"`
	SignerKey       string `json:"signer_key,omitempty"`
	SignerWeight    *int   `json:"signer_weight,omitempty"`

	SetFlags    []int    `json:"set_flags,omitempty"`
	SetFlagsS   []string `json:"set_flags_s,omitempty"`
	ClearFlags  []int    `json:"clear_flags,omitempty"`
	ClearFlagsS []string `json:"clear_flags_s,omitempty"`

	LowThreshold  *int `json:"low_threshold,omitempty"`
	MedThreshold  *int `json:"med_threshold,omitempty"`
	HighThreshold *int `json:"high_threshold,omitempty"`
}

// ChangeTrust is the json resource representing a single operation whose type
// is ChangeTrust.
type ChangeTrust struct {
	Base
	base.Asset
	Limit   string `json:"limit"`
	Trustee string `json:"trustee"`
	Trustor string `json:"trustor"`
}

// AllowTrust is the json resource representing a single operation whose type is
// AllowTrust.
type AllowTrust struct {
	Base
	base.Asset
	Trustee                        string `json:"trustee"`
	Trustor                        string `json:"trustor"`
	Authorize                      bool   `json:"authorize"`
	AuthorizeToMaintainLiabilities bool   `json:"authorize_to_maintain_liabilities"`
}

// AccountMerge is the json resource representing a single operation whose type
// is AccountMerge.
type AccountMerge struct {
	Base
	Account string `json:"account"`
	Into    string `json:"into"`
}

// Inflation is the json resource representing a single operation whose type is
// Inflation.
type Inflation struct {
	Base
}

// CreateClaimableBalance is the json resource representing a single operation whose type is
// CreateClaimableBalance.
type CreateClaimableBalance struct {
	Base
	Asset     string             `json:"asset"`
	Amount    string             `json:"amount"`
	Claimants []aurora.Claimant `json:"claimants"`
}

// ClaimClaimableBalance is the json resource representing a single operation whose type is
// ClaimClaimableBalance.
type ClaimClaimableBalance struct {
	Base
	BalanceID string `json:"balance_id"`
	Claimant  string `json:"claimant"`
}

// BeginSponsoringFutureReserves is the json resource representing a single operation whose type is
// BeginSponsoringFutureReserves.
type BeginSponsoringFutureReserves struct {
	Base
	SponsoredID string `json:"sponsored_id"`
}

// EndSponsoringFutureReserves is the json resource representing a single operation whose type is
// EndSponsoringFutureReserves.
type EndSponsoringFutureReserves struct {
	Base
	BeginSponsor string `json:"begin_sponsor,omitempty"`
}

// RevokeSponsorship is the json resource representing a single operation whose type is
// RevokeSponsorship.
type RevokeSponsorship struct {
	Base
	AccountID          *string `json:"account_id,omitempty"`
	ClaimableBalanceID *string `json:"claimable_balance_id,omitempty"`
	DataAccountID      *string `json:"data_account_id,omitempty"`
	DataName           *string `json:"data_name,omitempty"`
	OfferID            *int64  `json:"offer_id,omitempty,string"`
	TrustlineAccountID *string `json:"trustline_account_id,omitempty"`
	TrustlineAsset     *string `json:"trustline_asset,omitempty"`
	SignerAccountID    *string `json:"signer_account_id,omitempty"`
	SignerKey          *string `json:"signer_key,omitempty"`
}

// Operation interface contains methods implemented by the operation types
type Operation interface {
	PagingToken() string
	GetType() string
	GetID() string
	GetTransactionHash() string
	IsTransactionSuccessful() bool
}

// GetType returns the type of operation
func (base Base) GetType() string {
	return base.Type
}

// GetTypeI returns the ID of type of operation
func (base Base) GetTypeI() int32 {
	return base.TypeI
}

func (base Base) GetID() string {
	return base.ID
}

func (base Base) GetTransactionHash() string {
	return base.TransactionHash
}

func (base Base) IsTransactionSuccessful() bool {
	return base.TransactionSuccessful
}

// OperationsPage is the json resource representing a page of operations.
// OperationsPage.Record can contain various operation types.
type OperationsPage struct {
	Links    hal.Links `json:"_links"`
	Embedded struct {
		Records []Operation
	} `json:"_embedded"`
}

func (ops *OperationsPage) UnmarshalJSON(data []byte) error {
	var opsPage struct {
		Links    hal.Links `json:"_links"`
		Embedded struct {
			Records []interface{}
		} `json:"_embedded"`
	}

	if err := json.Unmarshal(data, &opsPage); err != nil {
		return err
	}

	for _, j := range opsPage.Embedded.Records {
		var b Base
		dataString, err := json.Marshal(j)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(dataString, &b); err != nil {
			return err
		}

		op, err := UnmarshalOperation(b.TypeI, dataString)
		if err != nil {
			return err
		}

		ops.Embedded.Records = append(ops.Embedded.Records, op)
	}

	ops.Links = opsPage.Links
	return nil
}

// UnmarshalOperation decodes responses to the correct operation struct
func UnmarshalOperation(operationTypeID int32, dataString []byte) (ops Operation, err error) {
	switch xdr.OperationType(operationTypeID) {
	case xdr.OperationTypeCreateAccount:
		var op CreateAccount
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypePathPaymentStrictReceive:
		var op PathPayment
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypePayment:
		var op Payment
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeManageSellOffer:
		var op ManageSellOffer
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeCreatePassiveSellOffer:
		var op CreatePassiveSellOffer
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeSetOptions:
		var op SetOptions
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeChangeTrust:
		var op ChangeTrust
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeAllowTrust:
		var op AllowTrust
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeAccountMerge:
		var op AccountMerge
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeInflation:
		var op Inflation
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeManageData:
		var op ManageData
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeBumpSequence:
		var op BumpSequence
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeManageBuyOffer:
		var op ManageBuyOffer
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypePathPaymentStrictSend:
		var op PathPaymentStrictSend
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeCreateClaimableBalance:
		var op CreateClaimableBalance
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeClaimClaimableBalance:
		var op ClaimClaimableBalance
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeBeginSponsoringFutureReserves:
		var op BeginSponsoringFutureReserves
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeEndSponsoringFutureReserves:
		var op EndSponsoringFutureReserves
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	case xdr.OperationTypeRevokeSponsorship:
		var op RevokeSponsorship
		if err = json.Unmarshal(dataString, &op); err != nil {
			return
		}
		ops = op
	default:
		err = errors.New("Invalid operation format, unable to unmarshal json response")
	}

	return
}
