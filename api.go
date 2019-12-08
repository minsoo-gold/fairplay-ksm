package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/Cooomma/ksm/d"
	"github.com/Cooomma/ksm/ksm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SpcMessage struct {
	Spc     string `json:"spc" binding:"required"`
	assetID string `json:"assetID"`
}

type ErrorMessage struct {
	Status  int    `json:"status" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type CkcResult struct {
	Ckc string `json:"ckc" binding:"required"`
}

type KeyRequest struct{ TicketID string }

type KeyResponse struct {
	Key string `json:"key" binding:"required"`
	IV  string `json:"iv" binding:"required"`
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      DefaultSkipper,
		AllowOrigins: []string{"*"}, // FIXME: Use your stremaing domain.
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPost},
	}))

	e.GET("/", handleGETHealth)
	e.GET("/healthy", handleGETHealth)
	e.POST("/fps/license", handlePOSTLicense)
	e.POST("/key", handlePOSTGenerateKey)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleGETHealth(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "OK")
}

func handlePOSTGenerateKey(ctx echo.Context) error {
	keyRequest := new(KeyRequest)
	if err := ctx.Bind(keyRequest); err != nil {
		errorMessage := &ErrorMessage{Status: 400, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, errorMessage)
	}
	return ctx.JSON(http.StatusOK, KeyResponse{
		Key: hex.EncodeToString(generateTicketKey([]byte(keyRequest.TicketID))),
		IV:  hex.EncodeToString(generateTicketIV([]byte(keyRequest.TicketID))),
	})
}

func handlePOSTLicense(ctx echo.Context) error {
	spcMessage := new(SpcMessage)
	if err := ctx.Bind(spcMessage); err != nil {
		errorMessage := &ErrorMessage{Status: 400, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, errorMessage)
	}

	fmt.Printf("%v\n", spcMessage)

	playback, err := base64.StdEncoding.DecodeString(spcMessage.Spc)
	checkError(err)

	if pub == "" {
		return ctx.JSON(http.StatusInternalServerError, "Certification Error.")
	}

	if pri == "" {
		return ctx.JSON(http.StatusInternalServerError, "Private Key Error.")
	}

	k := &ksm.Ksm{
		Pub:       pub,
		Pri:       pri,
		Rck:       RandomContentKey{}, // Use random content key for testing
		DFunction: d.AppleD{},         // Use D function provided by Apple Inc.
		Ask:       []byte{},
	}
	ckc, err2 := k.GenCKC(playback)
	checkError(err2)

	return ctx.JSON(http.StatusOK, CkcResult{
		Ckc: base64.StdEncoding.EncodeToString(ckc),
	})

}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// Random content key
type RandomContentKey struct {
}

// Implement FetchContentKey func
func (RandomContentKey) FetchContentKey(assetID []byte) ([]byte, []byte, error) {
	return generateTicketKey(assetID), generateTicketIV(assetID), nil
}

// Implement FetchContentKeyDuration func
func (RandomContentKey) FetchContentKeyDuration(assetID []byte) (*ksm.CkcContentKeyDurationBlock, error) {

	LeaseDuration := rand.Uint32()  // The duration of the lease, if any, in seconds.
	RentalDuration := rand.Uint32() // The duration of the rental, if any, in seconds.

	return ksm.NewCkcContentKeyDurationBlock(LeaseDuration, RentalDuration), nil
}

func envBase64Decode(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		fmt.Println("Env Decode Error: ", err)
		return ""
	}
	return string(data)
}

var pub = envBase64Decode(os.Getenv("FAIRPLAY_CERTIFICATION"))
var pri = envBase64Decode(os.Getenv("FAIRPLAY_PRIVATE_KEY"))
var ask = envBase64Decode(os.Getenv("FAIRPLAY_APPLICATION_SERVICE_KEY"))
