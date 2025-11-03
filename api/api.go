package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minsoo-gold/fairplay-ksm/cryptos"
	"github.com/minsoo-gold/fairplay-ksm/ksm"
	"github.com/minsoo-gold/fairplay-ksm/logger"
	"google.golang.org/api/option"
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

// Firestore 고객사 키 구조
type CustomerKeys struct {
	Certification string `firestore:"FAIRPLAY_CERTIFICATION"`
	PrivateKey    string `firestore:"FAIRPLAY_PRIVATE_KEY"`
	AppServiceKey string `firestore:"FAIRPLAY_APPLICATION_SERVICE_KEY"`
}

type CustomerKey struct {
	DocID         string `json:"doc_id"`
	Certification string `json:"FAIRPLAY_CERTIFICATION"`
	PrivateKey    string `json:"FAIRPLAY_PRIVATE_KEY"`
	AppServiceKey string `json:"FAIRPLAY_APPLICATION_SERVICE_KEY"`
}

type FairplayKeys struct {
	ClientID string `firestore:"client_id"`
	KID      string `firestore:"kid"`
	Key      string `firestore:"key"`
	IV       string `firestore:"iv"`
}

type FairplayKey struct {
	DocID    string `json:"doc_id" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
	KID      string `json:"kid" binding:"required"`
	Key      string `json:"key" binding:"required"`
	IV       string `json:"iv" binding:"required"`
}

// Firestore 클라이언트 전역
var firestoreClient *firestore.Client

func init() {
	// .env 파일 로드 (로컬 개발용)
	envPaths := []string{".env", "../.env", "../../.env"}
	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			loaded = true
			logger.Printf("Loaded .env from: %s", path)
			break
		}
	}
	if !loaded {
		logger.Println("No .env file found, relying on system environment variables")
	}

	// Firestore 초기화
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	// dbID := os.Getenv("FIRESTORE_DATABASE_ID")
	keyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if _, err := os.Stat(keyPath); err == nil {
		// 로컬 서비스 계정 키 사용
		client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(keyPath))
		if err != nil {
			panic(fmt.Sprintf("Failed to init Firestore client: %v", err))
		}
		firestoreClient = client
	} else {
		// Cloud Run 내장 서비스 계정 사용
		client, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			panic(fmt.Sprintf("Failed to init Firestore client: %v", err))
		}
		firestoreClient = client
	}
}

func getFairplayKeys(ctx context.Context, docID string) (*FairplayKeys, error) {
	doc, err := firestoreClient.Collection("fairplay").Doc(docID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var keys FairplayKeys
	if err := doc.DataTo(&keys); err != nil {
		return nil, err
	}
	return &keys, nil
}

func getCustomerKeys(ctx context.Context, docID string) (*CustomerKeys, error) {
	doc, err := firestoreClient.Collection("customer").Doc(docID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var keys CustomerKeys
	if err := doc.DataTo(&keys); err != nil {
		return nil, err
	}
	return &keys, nil
}

// Base64 Decode
func envBase64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		fmt.Println("Base64 Decode Error: ", err)
		return []byte{}
	}
	return data
}

func ReadPublicCert(keys *CustomerKeys) *rsa.PublicKey {
	pubEnvVar := envBase64Decode(keys.Certification)
	pubCert, err := cryptos.ParsePublicCertification(pubEnvVar)
	if err != nil {
		panic(err)
	}
	return pubCert
}

func ReadPriKey(keys *CustomerKeys) *rsa.PrivateKey {
	priEnvVar := envBase64Decode(keys.PrivateKey)
	priKey, err := cryptos.DecryptPriKey(priEnvVar, []byte("axissoft1@"))
	if err != nil {
		panic(err)
	}
	return priKey
}

func ReadASk(keys *CustomerKeys) []byte {
	ask, err := hex.DecodeString(keys.AppServiceKey)
	if err != nil {
		panic(err)
	}
	return ask
}

func main() {
	defer firestoreClient.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPost},
	}))

	e.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "KSM OK")
	})

	e.POST("/license", func(ctx echo.Context) error {
		client_id := ctx.QueryParam("client_id")
		if client_id == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "client_id query param required"})
		}

		//고객사 정보 가져오기
		customerKeys, err := getCustomerKeys(ctx.Request().Context(), client_id)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to load customer keys: %v", err)})
		}
		fmt.Println("getCustomerKeys:", customerKeys)

		// FirestoreContentKey 인스턴스 생성
		contentKey := NewFirestoreContentKey(ctx.Request().Context())

		k := &ksm.Ksm{
			Pub: ReadPublicCert(customerKeys),
			Pri: ReadPriKey(customerKeys),
			Rck: contentKey,
			Ask: ReadASk(customerKeys),
		}

		spcMessage := new(SpcMessage)
		contentType := ctx.Request().Header.Get("Content-Type")
		if err := ctx.Bind(spcMessage); err != nil {
			errorMessage := &ErrorMessage{Status: 400, Message: err.Error()}
			return ctx.JSON(http.StatusBadRequest, errorMessage)
		}

		var playback []byte
		var base64EncodingMethod string

		if strings.Contains(spcMessage.Spc, "-") || strings.Contains(spcMessage.Spc, "_") {
			base64EncodingMethod = "URL"
			playback, err = base64.URLEncoding.DecodeString(spcMessage.Spc)
		} else if strings.Contains(spcMessage.Spc, " ") && strings.Contains(spcMessage.Spc, "/") {
			base64EncodingMethod = "STD"
			playback, err = base64.StdEncoding.DecodeString(strings.ReplaceAll(spcMessage.Spc, " ", "+"))
		} else {
			base64EncodingMethod = "STD"
			playback, err = base64.StdEncoding.DecodeString(spcMessage.Spc)
		}

		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Failed to decode SPC: %v", err)})
		}

		ckc, err := k.GenCKC(playback)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to generate CKC: %v", err)})
		}

		var result string
		if base64EncodingMethod == "URL" {
			result = base64.URLEncoding.EncodeToString(ckc)
		} else {
			result = base64.StdEncoding.EncodeToString(ckc)
		}

		switch contentType {
		case "application/json":
			return ctx.JSON(http.StatusOK, &CkcResult{Ckc: result})
		default:
			return ctx.Blob(http.StatusOK, "application/x-www-form-urlencoded", []byte("<ckc>"+result+"</ckc>"))
		}
	})

	// customer 저장 API
	e.POST("/customer", func(ctx echo.Context) error {
		var c CustomerKey
		if err := ctx.Bind(&c); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Invalid request body: %v", err),
			})
		}

		// key도 직접 생성해서 사용 (tenant_id > ovp의 client_id)
		if c.DocID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "doc_id required",
			})
		}

		_, err := firestoreClient.Collection("customer").Doc(c.DocID).Set(ctx.Request().Context(), map[string]interface{}{
			"FAIRPLAY_CERTIFICATION":           c.Certification,
			"FAIRPLAY_PRIVATE_KEY":             c.PrivateKey,
			"FAIRPLAY_APPLICATION_SERVICE_KEY": c.AppServiceKey,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to save customer keys: %v", err),
			})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"status": "success",
			"doc_id": c.DocID,
		})
	})

	// fairplay 저장 API
	e.POST("/fairplay", func(ctx echo.Context) error {
		var fp FairplayKey
		if err := ctx.Bind(&fp); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Invalid request body: %v", err),
			})
		}

		// key도 직접 생성해서 사용 (asset_id)
		if fp.DocID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "doc_id required",
			})
		}

		_, err := firestoreClient.Collection("fairplay").Doc(fp.DocID).Set(ctx.Request().Context(), map[string]interface{}{
			"client_id": fp.ClientID,
			"kid":       fp.KID,
			"key":       fp.Key,
			"iv":        fp.IV,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to save fairplay key: %v", err),
			})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"status": "success",
			"doc_id": fp.DocID,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // 로컬 기본 포트
	}

	fmt.Printf("Starting server on port %s...\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}
