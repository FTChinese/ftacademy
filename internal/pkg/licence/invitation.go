package licence

import (
	"fmt"
	"github.com/FTChinese/ftacademy/internal/pkg"
	"github.com/FTChinese/ftacademy/internal/pkg/admin"
	"github.com/FTChinese/ftacademy/internal/pkg/ids"
	"github.com/FTChinese/ftacademy/internal/pkg/input"
	"github.com/FTChinese/ftacademy/internal/pkg/reader"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/rand"
	"github.com/guregu/null"
	"time"
)

// Invitation is an email sent to team member to accept a licence.
// An invitation could in 3 phases:
// Initially created: it indicates an email is sent to reader;
// Accepted: reader clicked the link in the invitation email,
// it should not be used any longer;
// Revoked: admin could revoke an invitation before it is accepted.
// An accepted invitation could not be revoked since that is meaningless.
// TODO: should we allow an invitation be re-sent if user failed to receive the email? Or just let admin to create a new invitation?
type Invitation struct {
	ID string `json:"id" db:"invite_id"`
	admin.Creator
	Status         InvitationStatus `json:"status" db:"invite_status"`
	Description    null.String      `json:"description" db:"invite_desc"`
	ExpirationDays int64            `json:"expirationDays" db:"invite_expiration_days"`
	Email          string           `json:"email" db:"invite_email"`
	LicenceID      string           `json:"licenceId" db:"licence_id"`
	Token          string           `json:"-" db:"invite_token"` // This field is used only when inserting data. Retrieval does not include this field. However, it is included when saving to the JSON column in licence.
	admin.RowTime
}

func NewInvitation(params input.InvitationParams, p admin.PassportClaims) (Invitation, error) {
	token, err := rand.Hex(32)
	if err != nil {
		return Invitation{}, err
	}

	return Invitation{
		ID: ids.InvitationID(),
		Creator: admin.Creator{
			AdminID: p.AdminID,
			TeamID:  p.TeamID.String,
		},
		Description:    params.Description,
		ExpirationDays: 7,
		Email:          params.Email,
		LicenceID:      params.LicenceID,
		Status:         InvitationStatusCreated,
		Token:          token,
		RowTime:        admin.NewRowTime(),
	}, nil
}

// IsExpired tests whether the invitation is expired.
// An expired invitation is not allowed grant its related licence.
func (i Invitation) IsExpired() bool {
	now := time.Now().Unix()

	created := i.CreatedUTC.Time.Unix()

	// Default 7 days * 24 * 60 * 60
	return (created + i.ExpirationDays*86400) < now
}

// IsAcceptable determines whether an invitation is valid.
// A valid invitation must be not expires, not revoked by admin, not accepted by any one.
// A valid invitation can be accepted or revoked.
func (i Invitation) IsAcceptable() bool {
	return i.Status == InvitationStatusCreated && !i.IsExpired()
}

// Accepted invalidates an invitation after reader accepted the licence associated with it.
func (i Invitation) Accepted() Invitation {
	i.Status = InvitationStatusAccepted
	i.UpdatedUTC = chrono.TimeNow()

	return i
}

func (i Invitation) IsRevocable() bool {
	return i.Status == InvitationStatusCreated
}

// Revoked invalidates an invitation by admin.
func (i Invitation) Revoked() Invitation {
	i.Status = InvitationStatusRevoked
	i.UpdatedUTC = chrono.TimeNow()

	return i
}

func (i Invitation) FormatDuration() string {
	return fmt.Sprintf("%d天", i.ExpirationDays)
}

// InvitationList is used for restful output.
type InvitationList struct {
	pkg.PagedList
	Data []Invitation `json:"data"`
}

type InvitationRevoked struct {
	Licence    ExpandedLicence `json:"licence"`
	Invitation Invitation      `json:"invitation"`
}

// InvitationVerified is returned after an invitation link
// is clicked and the corresponding ExpandedLicence is found.
type InvitationVerified struct {
	Licence    ExpandedLicence   `json:"licence"` // The licence being invited.
	Assignee   Assignee          `json:"assignee"`
	Membership reader.Membership `json:"membership"`
}
