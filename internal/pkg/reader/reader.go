package reader

import (
	"github.com/FTChinese/ftacademy/internal/pkg/licence"
	"github.com/guregu/null"
)

// Reader is a member of a team who will be granted
// a licence.
// When an invitee received an email, it should click the link
// in the email.
// When the link is opened, we take the following steps in order:
// 1. First check the if the token exists. A token is valid only
// when it does exist, is not accepted yet, and not revoked,
//and not expired.
// 2. Then we should check if the licence to be granted is
// still available.
// 3. Check if the invitee has an account at FTC. If not, ask it to sign up.
// 4. For existing user, check whether the account has a valid
// membership with it. It it does have one, deny the granting.
// 5. Link the licence to user's ftc id;
// 6. Insert membership if user does not have membership yet, or
// backup existing membership and update membership.
// 6. Mark the invitation as accepted;
type Reader struct {
	licence.Assignee
	Membership
}

// SignUp is used to create a new ftc user.
// Deprecated.
type SignUp struct {
	ID       string `db:"ftc_id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Token    string `db:"token"` // verification token
}

// Turn the SignUp into a new Reader type.
func (s SignUp) TeamMember(teamID string) licence.Staffer {
	return licence.Staffer{
		Email:  s.Email,
		FtcID:  null.StringFrom(s.ID),
		TeamID: teamID,
	}
}
