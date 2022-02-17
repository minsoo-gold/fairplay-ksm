package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cooomma/fairplay-ksm/cryptos"
	"github.com/cooomma/fairplay-ksm/ksm"
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

	k := &ksm.Ksm{
		Pub: ReadPublicCert(),
		Pri: ReadPriKey(),
		Rck: ksm.RandomContentKey{}, //NOTE: Don't use ramdom key in your application.
		Ask: ReadASk(),
	}

	e.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	e.POST("/fps/license", func(ctx echo.Context) error {
		spcMessage := new(SpcMessage)
		var playback []byte
		var base64EncodingMethod string
		contentType := ctx.Request().Header.Get("Content-Type")

		if err := ctx.Bind(spcMessage); err != nil {
			errorMessage := &ErrorMessage{Status: 400, Message: err.Error()}
			fmt.Println(ctx.Request().Header)
			fmt.Println(ctx.Request().Body)
			return ctx.JSON(http.StatusBadRequest, errorMessage)
		}

		if strings.Contains(spcMessage.Spc, "-") || strings.Contains(spcMessage.Spc, "_") {
			base64EncodingMethod = "URL"
			decoded, err := base64.URLEncoding.DecodeString(spcMessage.Spc)
			if err != nil {
				panic(err)
			}
			playback = decoded
		} else if strings.Contains(spcMessage.Spc, " ") && strings.Contains(spcMessage.Spc, "/") {
			base64EncodingMethod = "STD"
			decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(spcMessage.Spc, " ", "+"))
			if err != nil {
				panic(err)
			}
			playback = decoded
		} else {
			base64EncodingMethod = "STD"
			decoded, err := base64.StdEncoding.DecodeString(spcMessage.Spc)
			if err != nil {
				panic(err)
			}
			playback = decoded
		}

		ckc, err := k.GenCKC(playback)
		if err != nil {
			panic(err)
		}

		var result string

		switch base64EncodingMethod {
		case "URL":
			result = base64.URLEncoding.EncodeToString(ckc)
		case "STD":
			result = base64.StdEncoding.EncodeToString(ckc)
		default:
			result = base64.StdEncoding.EncodeToString(ckc)
		}

		fmt.Println(result)

		switch contentType {
		case "application/json":
			return ctx.JSON(200, &CkcResult{Ckc: result})
		case "application/x-www-form-urlencoded":
			return ctx.Blob(200, "application/x-www-form-urlencoded", []byte("<ckc>"+result+"</ckc>"))
		default:
			return ctx.Blob(200, "application/x-www-form-urlencoded", []byte("<ckc>"+result+"</ckc>"))
		}
	})
	e.Logger.Fatal(e.Start(":8080"))
}

func envBase64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		fmt.Println("Env Decode Error: ", err)
		return []byte{}
	}
	return data
}

func ReadPublicCert() *rsa.PublicKey {
	pubEnvVar := envBase64Decode(os.Getenv("FAIRPLAY_CERTIFICATION"))

	if len(pubEnvVar) == 0 {
		panic("Can't not find FAIRPLAY CERTIFICATION")
	}

	pubCert, err := cryptos.ParsePublicCertification(pubEnvVar)
	if err != nil {
		panic(err)
	}
	return pubCert
}

func ReadPriKey() *rsa.PrivateKey {
	priEnvVar := envBase64Decode(os.Getenv("FAIRPLAY_PRIVATE_KEY"))

	if len(priEnvVar) == 0 {
		panic("Can't not find FAIRPLAY PRIVATE KEY")
	}
	priKey, err := cryptos.DecryptPriKey(priEnvVar, []byte(""))
	if err != nil {
		panic(err)
	}
	return priKey
}

func ReadASk() []byte {
	askEnvVar := os.Getenv("FAIRPLAY_APPLICATION_SERVICE_KEY")
	if len(askEnvVar) == 0 {
		askEnvVar = "d87ce7a26081de2e8eb8acef3a6dc179" //Apple provided
	}
	ask, err := hex.DecodeString(askEnvVar)
	if err != nil {
		panic(err)
	}
	return ask
}
