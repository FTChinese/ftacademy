package order

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/ftacademy/internal/pkg"
	"github.com/FTChinese/ftacademy/internal/pkg/admin"
	"github.com/FTChinese/ftacademy/pkg/price"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
)

// CheckoutProduct describes the quantity of a product put into an order.
// This is used when save all items of an order as JSON
// in an order's row.
type CheckoutProduct struct {
	Price         price.Price `json:"price"`
	NewCopies     int64       `json:"newCopies"`     // How many new copies user purchased
	RenewalCopies int64       `json:"renewalCopies"` // How many renewals user purchased.
}

// CheckoutProducts is used the retrieve/save an array of
// CheckoutProduct into db.
type CheckoutProducts []CheckoutProduct

// Value implements Valuer interface when saving
func (b CheckoutProducts) Value() (driver.Value, error) {
	j, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return string(j), nil
}

func (b *CheckoutProducts) Scan(src interface{}) error {
	if src == nil {
		*b = []CheckoutProduct{}
		return nil
	}
	switch s := src.(type) {
	case []byte:
		var tmp []CheckoutProduct
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*b = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to []CheckoutProduct")
	}
}

type BaseOrder struct {
	ID            string      `json:"id" db:"order_id"`
	AmountPayable float64     `json:"amountPayable" db:"amount_payable"`
	CreatedBy     string      `json:"createdBy" db:"created_by"`
	CreatedUTC    chrono.Time `json:"createdUtc" db:"created_utc"`
	ItemCount     int64       `json:"itemCount" db:"item_count"`
	Status        Status      `json:"status" db:"current_status"`
	TeamID        string      `json:"teamId" db:"team_id"`
}

// CheckoutOrder describes the details of each transaction
// to purchase a licence.
// If a transaction is used to purchase a new licence, the
// licence should be created together with the order but marked
// as inactive. Once the transaction is confirmed,
// the licence will be activated and the admin is allowed to
// invite someone to use this licence.
// If a transaction is used to renew/upgrade a licence,
// the licence associated with it won't be touched until
// it is confirmed, which will result licence extended or
// upgraded and the membership (if the licence is granted
// to someone) will be backed up and updated corresponding.
type CheckoutOrder struct {
	BaseOrder
	// An array of products, together with the quantities, use is trying to purchase.
	Products CheckoutProducts `json:"products" db:"checkout_products"`
}

func NewBriefOrder(cart ShoppingCart, p admin.PassportClaims) CheckoutOrder {
	return CheckoutOrder{
		BaseOrder: BaseOrder{
			ID:            pkg.OrderID(),
			AmountPayable: cart.TotalAmount,
			CreatedBy:     p.AdminID,
			CreatedUTC:    chrono.TimeNow(),
			ItemCount:     cart.ItemCount,
			Status:        StatusPending,
			TeamID:        p.TeamID.String,
		},
		Products: cart.ProductsBrief(),
	}
}

// Payment describes the details the an order's payment.
type Payment struct {
	AmountPaid    null.Float    `json:"amountPaid" db:"amount_paid"`
	ApprovedBy    null.String   `json:"approvedBy" db:"approved_by"`
	ApprovedUTC   chrono.Time   `json:"approvedUtc" db:"approved_utc"`
	Description   null.String   `json:"description" db:"description"`
	PaymentMethod PaymentMethod `json:"paymentMethod" db:"payment_method"`
	TransactionID null.String   `json:"transactionId" db:"transaction_id"`
}

// CheckoutOrderList contains a list of orders
type CheckoutOrderList struct {
	Total int64 `json:"total"`
	gorest.Pagination
	Data []CheckoutOrder `json:"data"`
	Err  error           `json:"-"`
}

// Order contains all details of what user wanted to buy,
// how payment is handled.
type Order struct {
	BaseOrder
	CartItems []CartItemSchema `json:"cartItems"`
	Payment
}
