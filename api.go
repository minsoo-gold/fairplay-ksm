package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/Cooomma/ksm/ksm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SpcMessage struct {
	Spc     string `json:"spc" binding:"required"`
	AssetID string `json:"assetID"`
}

type ErrorMessage struct {
	Status  int    `json:"status" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type CkcResult struct {
	Ckc string `json:"ckc" binding:"required"`
}

var FairplayPrivateKey = ReadPriKey()
var FairplayPublicCertification = ReadPublicCert()
var FairplayASk = ReadASk()

func main() {
	// Setting Echo
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // FIXME: Use your streaming domain.
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPost},
	}))

	e.GET("/", handleGETHealth)
	e.POST("/fps/license", handlePOSTLicense)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleGETHealth(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "OK")
}

func handlePOSTLicense(ctx echo.Context) error {
	spcMessage := new(SpcMessage)
	if err := ctx.Bind(spcMessage); err != nil {
		errorMessage := &ErrorMessage{Status: 400, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, errorMessage)
	}

	fmt.Printf("%v\n", spcMessage)

	playback, err := base64.StdEncoding.DecodeString(spcMessage.Spc)
	if err != nil {
		panic(err)
	}

	k := &ksm.Ksm{
		Pub: FairplayPublicCertification,
		Pri: FairplayPrivateKey,
		Rck: RandomContentKey{}, //FIXME: Use random content key for testing
		Ask: FairplayASk,
	}
	ckc, err := k.GenCKC(playback)
	if err != nil {
		panic(err)
	}

	return ctx.JSON(http.StatusOK, CkcResult{
		Ckc: base64.StdEncoding.EncodeToString(ckc),
	})

}
