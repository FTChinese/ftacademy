package controller

import (
	"github.com/FTChinese/ftacademy/internal/app/b2b/repository/api"
	"github.com/FTChinese/ftacademy/internal/pkg/reader"
	"github.com/FTChinese/ftacademy/pkg/config"
	"github.com/FTChinese/go-rest/render"
	"github.com/labstack/echo/v4"
	"log"
)

type ReaderRouter struct {
	guard     reader.JWTGuard
	apiClient api.Client
	version   string
}

func NewReaderRouter(client api.Client, appKey config.AppKey, version string) ReaderRouter {
	return ReaderRouter{
		guard:     reader.NewJWTGuard(appKey.GetJWTKey()),
		apiClient: client,
		version:   version,
	}
}

func (router ReaderRouter) RequireLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		claims, err := router.guard.RetrievePassportClaims(c.Request())
		if err != nil {
			log.Printf("Error parsing JWT %v", err)
			return render.NewUnauthorized(err.Error())
		}

		c.Set(claimsCtxKey, claims)
		return next(c)
	}
}