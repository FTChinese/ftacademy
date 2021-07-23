package pkg

const (
	SiteBaseURL        = "https://next.ftacademy.cn"
	B2BBaseURL         = SiteBaseURL + "/corporate"
	UserBaseURL        = SiteBaseURL + "/user"
	ReaderVerification = UserBaseURL + "/verification"
)

func B2BPasswordResetURL(token string) string {
	return B2BBaseURL + "/password-reset/" + token
}

func B2BVerifyAdminURL(token string) string {
	return B2BBaseURL + "/verify/" + token
}

func B2BVerifyInvitationURL(token string) string {
	return B2BBaseURL + "/verify-invitation/" + token
}