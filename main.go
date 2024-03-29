package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/FTChinese/ftacademy/internal/access"
	"github.com/FTChinese/ftacademy/internal/api"
	"github.com/FTChinese/ftacademy/internal/app/b2b"
	"github.com/FTChinese/ftacademy/internal/app/content"
	"github.com/FTChinese/ftacademy/internal/app/reader"
	"github.com/FTChinese/ftacademy/pkg/config"
	"github.com/FTChinese/ftacademy/pkg/db"
	"github.com/FTChinese/ftacademy/pkg/postman"
	"github.com/FTChinese/ftacademy/pkg/xhttp"
	"github.com/FTChinese/ftacademy/web"
	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed build/api.toml
var tomlConfig string

//go:embed build/version
var version string

//go:embed build/build_time
var build string

//go:embed client_version_b2b
var clientVersionB2B string

//go:embed client_version_reader
var clientVersionReader string

var (
	production bool
	liveMode   bool
)

func newFooter(cv string) web.Footer {
	return web.Footer{
		Year:          time.Now().Year(),
		ClientVersion: cv,
		ServerVersion: version,
	}
}

func init() {
	flag.BoolVar(&production, "production", true, "Connect to production MySQL database if true, or localhost if false")
	flag.BoolVar(&liveMode, "livemode", true, "Determine live/sandbox mode for Stripe")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	config.MustSetupViper([]byte(tomlConfig))
}

func main() {

	logger := config.MustGetLogger(production)

	myDBs := db.MustNewMyDBs()

	pm := postman.New(config.MustGetHanqiConn())

	//b2bGuard := controller.NewJWTGuard(b2bAppKey.GetJWTKey())
	oauthGuard := access.NewGuard(myDBs)

	apiClients := api.NewClients(production)

	adminRouter := b2b.NewAdminRouter(myDBs, pm, logger)
	subsRouter := b2b.NewSubsRouter(myDBs, pm, logger)
	productRouter := b2b.NewProductRouter(apiClients, logger)
	readerRouter := reader.NewReaderRouter(apiClients, version)
	stripeRouter := reader.NewStripeRouter(
		apiClients,
		production,
		logger)
	cmsRouter := b2b.NewCMSRouter(myDBs, pm, logger)
	legalRoutes := content.NewRoutes(
		apiClients.Select(true),
		version,
		logger)

	e := echo.New()
	e.Renderer = web.MustNewRenderer(!production)

	if !production {
		e.Static("/static", "build/public/static")
	}

	e.Pre(middleware.AddTrailingSlash())

	e.HTTPErrorHandler = web.ErrorHandler

	e.Use(xhttp.DumpRequest)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.CSRF())

	e.GET("/corporate/*", func(c echo.Context) error {
		return c.Render(http.StatusOK, "b2b/home.html", pongo2.Context{
			"footer": newFooter(clientVersionB2B),
		})
	}, xhttp.NoCache)

	e.GET("/reader/*", func(c echo.Context) error {
		return c.Render(http.StatusOK, "reader/home.html", pongo2.Context{
			"footer": newFooter(clientVersionReader),
		})
	}, xhttp.NoCache)

	legalDocGroup := e.Group("/terms")
	{
		legalDocGroup.GET("/", legalRoutes.ListLegalDoc)
		// Optional query parameter ?refresh=true
		legalDocGroup.GET("/:id/", legalRoutes.LoadLegalDoc)
	}

	serviceGroup := e.Group("/service")
	{
		serviceGroup.GET("/qr/", reader.GenerateQRImage)
	}

	apiGroup := e.Group("/api")

	// ---------------------------------------------
	// Reader section is restricted to FTC user only.
	// ---------------------------------------------
	readerAPIGroup := apiGroup.Group("/reader")
	readerAuthGroup := readerAPIGroup.Group("/auth")

	emailAuthGroup := readerAuthGroup.Group("/email")
	{
		emailAuthGroup.GET("/exists/", readerRouter.EmailExists)
		emailAuthGroup.POST("/login/", readerRouter.EmailLogin)
		emailAuthGroup.POST("/signup/", readerRouter.EmailSignUp)
		emailAuthGroup.POST("/verification/:token/", readerRouter.VerifyEmail)
	}
	mobileAuthGroup := readerAuthGroup.Group("/mobile")
	{
		mobileAuthGroup.PUT("/verification/", readerRouter.RequestMobileLoginSMS)
		mobileAuthGroup.POST("/verification/", readerRouter.VerifyMobileLoginSMS)
		mobileAuthGroup.POST("/link/", readerRouter.MobileLinkExistingEmail)
		mobileAuthGroup.POST("/signup/", readerRouter.MobileSignUp)
	}
	passwordResetGroup := readerAuthGroup.Group("/password-reset")
	{
		passwordResetGroup.POST("/", readerRouter.ResetPassword)
		passwordResetGroup.POST("/letter/", readerRouter.RequestPwResetLetter)
		passwordResetGroup.GET("/tokens/:token/", readerRouter.VerifyPwResetToken)
	}
	wxAuthGroup := readerAuthGroup.Group("/wx")
	{
		wxAuthGroup.GET("/code/", readerRouter.WxRequestCode)
		wxAuthGroup.POST("/login/", readerRouter.WxLogin)
		wxAuthGroup.POST("/refresh/", readerRouter.WxRefresh)
	}

	readerAccountGroup := readerAPIGroup.Group("/account", readerRouter.RequireLoggedIn)
	{
		readerAccountGroup.GET("/", readerRouter.LoadAccount)
		readerAccountGroup.GET("/jwt/", readerRouter.LoadAccountWithJWT)
		readerAccountGroup.PATCH("/email/", readerRouter.UpdateEmail)
		readerAccountGroup.POST("/request-verification/", readerRouter.RequestVerification)
		readerAccountGroup.PATCH("/name/", readerRouter.UpdateName)
		readerAccountGroup.PATCH("/password/", readerRouter.UpdatePassword)
		readerAccountGroup.PATCH("/mobile/", readerRouter.UpdateMobile)
		readerAccountGroup.PUT("/mobile/verification/", readerRouter.RequestMobileUpdateSMS)
		readerAccountGroup.GET("/address/", readerRouter.LoadAddress)
		readerAccountGroup.PATCH("/address/", readerRouter.UpdateAddress)
		readerAccountGroup.GET("/profile/", readerRouter.LoadProfile)
		readerAccountGroup.PATCH("/profile/", readerRouter.UpdateProfile)
		readerAccountGroup.POST("/wx/signup/", readerRouter.WxSignUp)
		readerAccountGroup.POST("/wx/link/", readerRouter.WxLink)
		readerAccountGroup.POST("/wx/unlink/", readerRouter.WxUnlink)
	}

	memberGroup := readerAPIGroup.Group("/membership", readerRouter.RequireLoggedIn)
	{
		memberGroup.GET("/", readerRouter.LoadMembership)
		memberGroup.POST("/addons/", readerRouter.ClaimAddon)
	}

	iapGroup := readerAPIGroup.Group("/apple", readerRouter.RequireLoggedIn)
	{
		iapGroup.POST("/subs/:id/", readerRouter.RefreshIAP)
	}

	// --------------------------
	// Paywall section is public.
	// --------------------------
	paywallGroup := apiGroup.Group("/paywall")
	{
		// All the following routes have query parameter:
		// - live=<boolean>, default to true
		paywallGroup.GET("/", productRouter.Paywall)
		paywallGroup.GET("/stripe/prices/:id/", productRouter.StripePrice)
		// If client passed `?live=true` via url query, the test publishable key will be returned.
		// When client is using test key, it must also use a test account so that the serve
		// knows it should send request to sandbox api.
		// By default, all requests are forwarded to production live server.
		// All endpoints requiring logged-in state get live/test mode from Json Web Token's claim field.
		paywallGroup.GET("/stripe/publishable-key/", stripeRouter.PublishableKey)
	}

	// For those endpoints with RequiredLoggedIn middleware, the JWT passport claims contains
	// Live field. You do not need to send `?live=true` in url since that would be vulnerable to cheating.
	// The passport claim is generate upon user login and always verified for each request.
	// We could assume this is a much safer approach for critical section like payment.
	ftcPayGroup := readerAPIGroup.Group("/ftc-pay", readerRouter.RequireLoggedIn)
	{
		ftcPayGroup.POST("/ali/desktop/", readerRouter.CreateAliOrder)
		ftcPayGroup.POST("/wx/desktop/", readerRouter.CreateWxOrder)
		ftcPayGroup.POST("/orders/:id/verify/", readerRouter.VerifyFtcOrder)
		ftcPayGroup.GET("/discounts/:id/", readerRouter.LoadOneOffDiscount)
	}

	stripeGroup := readerAPIGroup.Group("/stripe", readerRouter.RequireLoggedIn)
	{
		customerGroup := stripeGroup.Group("/customers")
		customerGroup.POST("/", stripeRouter.CreateCustomer)
		customerGroup.GET("/:id/", stripeRouter.GetCustomer)
		customerGroup.GET("/:id/default-payment-method/", stripeRouter.GetCusDefaultPaymentMethod)
		customerGroup.POST("/:id/default-payment-method/", stripeRouter.SetCusDefaultPaymentMethod)
		customerGroup.GET("/:id/payment-methods/", stripeRouter.ListCusPaymentMethods)

		setupGroup := stripeGroup.Group("/setup-intents")
		setupGroup.POST("/", stripeRouter.CreateSetupIntent)
		// ?refresh=true
		setupGroup.GET("/:id/", stripeRouter.GetSetupIntent)
		setupGroup.GET("/:id/payment-method/", stripeRouter.GetSetupPaymentMethod)

		pmGroup := stripeGroup.Group("/payment-methods")
		pmGroup.GET("/:id/", stripeRouter.GetPaymentMethod)

		subsGroup := stripeGroup.Group("/subs")
		subsGroup.POST("/", stripeRouter.CreateSubs)
		subsGroup.GET("/:id/", stripeRouter.GetSubs)
		subsGroup.POST("/:id/", stripeRouter.UpdateSubs)
		subsGroup.POST("/:id/refresh/", stripeRouter.RefreshSubs)
		subsGroup.POST("/:id/cancel/", stripeRouter.CancelSubs)
		subsGroup.POST("/:id/reactivate/", stripeRouter.ReactivateSubs)
		subsGroup.GET("/:id/default-payment-method/", stripeRouter.GetSubsDefaultPaymentMethod)
		subsGroup.POST("/:id/default-payment-method/", stripeRouter.UpdateSubsDefaultPaymentMethod)
		subsGroup.GET("/:id/latest-invoice/", stripeRouter.GetLatestInvoice)
		subsGroup.GET("/:id/latest-invoice/any-coupon/", stripeRouter.CouponOfLatestInvoice)
	}

	// -------------------------------------------------
	// B2B section is restricted to corporate only.
	// ------------------------------------------------
	b2bAPIGroup := apiGroup.Group("/b2b")

	b2bAuthGroup := b2bAPIGroup.Group("/auth")
	{
		b2bAuthGroup.POST("/login/", adminRouter.Login)
		b2bAuthGroup.POST("/signup/", adminRouter.SignUp)
		b2bAuthGroup.GET("/verify/:token/", adminRouter.VerifyEmail)

		pwResetGroup := b2bAuthGroup.Group("/password-reset")
		{
			// Handle resetting password
			pwResetGroup.POST("/", adminRouter.ResetPassword)

			// Sending forgot-password email
			pwResetGroup.POST("/letter/", adminRouter.ForgotPassword)

			// Verify forgot-password token.
			// If valid, redirect to /forgot-password.
			// If invalid, redirect to /forgot-password/letter to ask
			// user to enter email again.
			pwResetGroup.GET("/token/:token/", adminRouter.VerifyResetToken)
		}
	}

	b2bAccountGroup := b2bAPIGroup.Group("/account", adminRouter.RequireLoggedIn)
	{
		//b2bAccountGroup.GET("/", accountRouter.Account)
		b2bAccountGroup.GET("/jwt/", adminRouter.RefreshJWT)
		b2bAccountGroup.POST("/request-verification/", adminRouter.RequestVerification)
		b2bAccountGroup.PATCH("/display-name/", adminRouter.ChangeName)
		b2bAccountGroup.PATCH("/password/", adminRouter.ChangePassword)
	}

	b2bTeamGroup := b2bAPIGroup.Group("/team", adminRouter.RequireLoggedIn)
	{
		b2bTeamGroup.GET("/", adminRouter.LoadTeam)
		b2bTeamGroup.POST("/", adminRouter.CreateTeam)
		b2bTeamGroup.PATCH("/", adminRouter.UpdateTeam)
	}

	b2bSearchGroup := b2bAPIGroup.Group("/search", adminRouter.RequireLoggedIn)
	{
		// ?email=<string>
		b2bSearchGroup.GET("/membership/", subsRouter.FindMembership)
	}

	orderGroup := b2bAPIGroup.Group("/orders", adminRouter.RequireTeamSet)
	{
		// List orders
		orderGroup.GET("/", subsRouter.ListOrders)
		// CreateTeam orders, or renew/upgrade in bulk.
		orderGroup.POST("/", subsRouter.CreateOrders)
		orderGroup.GET("/:id/", subsRouter.LoadOrder)
	}

	b2bLicenceGroup := b2bAPIGroup.Group("/licences", adminRouter.RequireTeamSet)
	{
		// List licences
		b2bLicenceGroup.GET("/", subsRouter.ListLicence)
		b2bLicenceGroup.GET("/:id/", subsRouter.LoadLicence)
		// Revoked a licence
		b2bLicenceGroup.POST("/:id/revoke/", subsRouter.RevokeLicence)
	}

	b2bInvitationGroup := b2bAPIGroup.Group("/invitations", adminRouter.RequireTeamSet)
	{
		// List invitations
		b2bInvitationGroup.GET("/", subsRouter.ListInvitations)
		// Create invitation.
		// Also update the linked licence's status.
		b2bInvitationGroup.POST("/", subsRouter.CreateInvitation)
		// Revoked invitation before licence is accepted.
		// Also revert the status of a licence from invitation sent
		// back to available.
		b2bInvitationGroup.POST("/:id/revoke/", subsRouter.RevokeInvitation)
	}

	// Steps to accept an invitation:
	// 1. Open token url and the token is valid;
	// 2. Use email to find user account (If account not found, go to signup);
	// 3. Get account data and find out if membership already exists
	// 4. Grant licence
	b2bGrantGroup := b2bAPIGroup.Group("/licence")
	{
		// Verify the invitation is valid.
		b2bGrantGroup.GET("/invitation/verification/:token/", subsRouter.VerifyInvitation)

		// Grant licence to user
		b2bGrantGroup.POST("/grant/", subsRouter.GrantLicence)
	}

	//-------------------------------------------------
	// The following is used by internal system.
	// It is not used by any customer-side client.
	// Instead, it is a restful API used by superyard
	// to forward request since I do not want to repeat
	// identical data type definition and manipulation
	// inside another Golang app.
	// -----------------------------------------------
	cmsGroup := apiGroup.Group("/cms", oauthGuard.RequireToken)
	{
		cmsGroup.GET("/profile/:id/", cmsRouter.LoadingAdminProfile)
		// List teams
		//cmsGroup.GET("/teams/",)
		// Show team detail
		// * admin account;
		// * team name
		// * orders
		// * licences
		cmsGroup.GET("/teams/:id/", cmsRouter.LoadTeam)
		// List orders
		// Query parameters used as filters:
		// team=xxx - List orders of the specified team
		// status=pending_payment | paid | processing | cancelled - List orders of the specified status
		cmsGroup.GET("/orders/", cmsRouter.ListOrders)
		// Details of an order:
		// * order data
		// * team details
		cmsGroup.GET("/orders/:id/", cmsRouter.LoadOrder)
		// Order payment confirmed.
		cmsGroup.POST("/orders/:id/", cmsRouter.ConfirmPayment)
	}

	e.Logger.Fatal(e.Start(":4000"))
}
