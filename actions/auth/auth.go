package auth

import (
	"errors"
	"idp/models/crypto"
	"idp/models/db"
	"idp/models/jwt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

type verify struct {
	Msg  string `json:"msg" form:"msg"`
	Sig  string `json:"sig" form:"sig" validate:"required"`
	Addr string `json:"addr" form:"addr"`
}

// Verify response signed jwt token, if pass
func Verify(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.SetCookie(&http.Cookie{
				Name:     "IDHUB_JWT",
				HttpOnly: true,
				Path:     "/",
				MaxAge:   -1,
			})

			c.SetCookie(&http.Cookie{
				Name:     "IDHUB_IDENTITY",
				HttpOnly: false,
				Path:     "/",
				MaxAge:   -1,
			})

			err = c.String(http.StatusUnauthorized, r.(error).Error())
		}
	}()

	v := new(verify)
	c.Bind(v)

	err = c.Validate(v)

	if err != nil {
		panic(err)
	}

	msg, err := db.GetVerifyMsg(v.Addr)

	addr, err := crypto.EcRecover(msg, v.Sig)

	if err != nil {
		panic(err)
	}

	if strings.ToLower(addr) != strings.ToLower(v.Addr) {
		panic(errors.New("verify failed"))
	} else {
		addr = v.Addr
	}

	tokenString, err := jwt.Sign(map[string]interface{}{
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(30 * time.Minute).Unix(),
		"iss":      "IDHub IdP",
		"sub":      "IDHub identity is all your life",
		"identity": addr,
	})

	if err != nil {
		panic(err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "IDHUB_JWT",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
	})

	c.SetCookie(&http.Cookie{
		Name:     "IDHUB_IDENTITY",
		Value:    addr,
		HttpOnly: false,
		Path:     "/",
		Expires:  time.Now().Add(30 * time.Minute),
	})

	// return c.String(http.StatusOK, addr)
	return c.NoContent(http.StatusOK)
}

// Booking response the register message
func Booking(c echo.Context) error {
	v := &verify{}
	c.Bind(v)

	msg, err := db.GetBookingMsg(v.Addr)

	if err != nil {
		return c.String(http.StatusNotAcceptable, err.Error())
	}

	return c.String(http.StatusOK, msg)
}

// Logout will delete the cookie
func Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "IDHUB_JWT",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	c.SetCookie(&http.Cookie{
		Name:     "IDHUB_IDENTITY",
		HttpOnly: false,
		Path:     "/",
		MaxAge:   -1,
	})

	return c.NoContent(http.StatusOK)
}
