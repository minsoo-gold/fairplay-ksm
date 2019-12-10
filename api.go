package main

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/Cooomma/ksm/crypto"
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

	// Setting Logger
	// log.SetFormatter(&log.JSONFormatter{})
	// log.SetOutput(os.Stdout)
	// log.SetLevel(log.DebugLevel)

	// Setting Echo
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // FIXME: Use your stremaing domain.
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPost},
	}))

	e.GET("/", handleGETHealth)
	e.GET("/healthy", handleGETHealth)
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
	checkError(err)

	pubCert, err := readPublicCert()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Public Certification Error")
	}

	priKey, err := readPriKey()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Private Key Error")
	}

	ask, err := readASk()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "ASk Error")
	}

	k := &ksm.Ksm{
		Pub: pubCert,
		Pri: priKey,
		Rck: RandomContentKey{}, // Use random content key for testing
		//DFunction: DFunction{},        // Use D function provided by Apple Inc.
		Ask: ask,
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
	return generateDummyKeyIVPair(assetID)
}

// Implement FetchContentKeyDuration func
func (RandomContentKey) FetchContentKeyDuration(assetID []byte) (*ksm.CkcContentKeyDurationBlock, error) {

	LeaseDuration := rand.Uint32()  // The duration of the lease, if any, in seconds.
	RentalDuration := rand.Uint32() // The duration of the rental, if any, in seconds.

	return ksm.NewCkcContentKeyDurationBlock(LeaseDuration, RentalDuration), nil
}

func generateDummyKeyIVPair(assetID []byte) ([]byte, []byte, error) {
	dummyKey := make([]byte, 16)
	dummyIV := make([]byte, 16)
	rand.Read(dummyIV)

	if len(assetID) == 0 {
		rand.Read(dummyKey)
		return dummyKey, dummyIV, nil
	}
	// NOTE: Here is to implement your key generator.
	generator := md5.New()
	generator.Write(assetID)
	dummyKey = generator.Sum(nil)
	return dummyKey, dummyIV, nil
}

func envBase64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		fmt.Println("Env Decode Error: ", err)
		return []byte{}
	}
	return data
}

func readPublicCert() (*rsa.PublicKey, error) {
	pubEnvVar := envBase64Decode(os.Getenv("FAIRPLAY_CERTIFICATION"))

	if len(pubEnvVar) == 0 {
		panic("Can't not find FAIRPLAY CERTIFICATION")
	}
	pubCert, err := crypto.ParsePublicCertification(pubEnvVar)
	if err != nil {
		return nil, err
	}
	return pubCert, nil
}

func readPriKey() (*rsa.PrivateKey, error) {
	priEnvVar := envBase64Decode(os.Getenv("FAIRPLAY_PRIVATE_KEY"))

	if len(priEnvVar) == 0 {
		panic("Can't not find FAIRPLAY PRIVATE KEY")
	}
	priKey, err := crypto.DecryptPriKey(priEnvVar, []byte(""))
	if err != nil {
		return nil, err
	}
	return priKey, nil
}

func readASk() ([]byte, error) {
	askEnvVar := os.Getenv("FAIRPLAY_APPLICATION_SERVICE_KEY")
	if len(askEnvVar) == 0 {
		askEnvVar = "d87ce7a26081de2e8eb8acef3a6dc179" //Apple provided
	}
	ask, _ := hex.DecodeString(askEnvVar)
	return ask, nil
}
