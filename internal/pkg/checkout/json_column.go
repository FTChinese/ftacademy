package checkout

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/FTChinese/ftacademy/internal/pkg/input"
	"github.com/FTChinese/ftacademy/internal/pkg/licence"
)

// ExpandedLicenceJSON is used to implement sql.Valuer interface
// so that we could save it directly as JSON.
type ExpandedLicenceJSON struct {
	licence.ExpandedLicence
}

func (l ExpandedLicenceJSON) Value() (driver.Value, error) {
	// Return NULL for zero value.
	if l.ID == "" {
		return nil, nil
	}

	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (l *ExpandedLicenceJSON) Scan(src interface{}) error {
	if src == nil {
		*l = ExpandedLicenceJSON{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp ExpandedLicenceJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to ExpandedLicenceJSON")
	}
}

// ExpLicenceListJSON is used to save a list of ExpandedLicence as JSON.
// when saving a CartItem
type ExpLicenceListJSON []licence.ExpandedLicence

func (l ExpLicenceListJSON) Value() (driver.Value, error) {
	if len(l) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (l *ExpLicenceListJSON) Scan(src interface{}) error {
	if src == nil {
		*l = ExpLicenceListJSON{}
		return nil
	}
	switch s := src.(type) {
	case []byte:
		var tmp []licence.ExpandedLicence
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to ExpLicenceListJSON")
	}
}

type TeamJSON struct {
	input.TeamParams
}

func (t TeamJSON) Value() (driver.Value, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (t *TeamJSON) Scan(src interface{}) error {
	if src == nil {
		*t = TeamJSON{}
	}
	switch s := src.(type) {
	case []byte:
		var tmp TeamJSON
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}

		*t = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to TeamJSON")
	}
}

// OrderItemListJSON is used to save a list of OrderItem.
// It corresponds to the order.items_summary field
// in DB so that when retrieving data, we don't need to
// retrieve all details of an order's items.
type OrderItemListJSON []OrderItem

func (l OrderItemListJSON) Value() (driver.Value, error) {
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (l *OrderItemListJSON) Scan(src interface{}) error {
	if src == nil {
		*l = []OrderItem{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		var tmp []OrderItem
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to []OrderItem")
	}
}
