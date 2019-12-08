module github.com/Coooomma/ksm

go 1.13

replace (
	github.com/Cooomma/crypto => ./crypto
	github.com/Cooomma/crypto/aes => ./crypto/aes
	github.com/Cooomma/crypto/rsa => ./crypto/rsa
	github.com/Cooomma/ksm/d => ./d
	github.com/Cooomma/ksm/ksm => ./ksm
	github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2
)

require (
	github.com/Cooomma/crypto v0.0.0
	github.com/Cooomma/crypto/aes v0.0.0
	github.com/Cooomma/crypto/rsa v0.0.0
	github.com/Cooomma/ksm/d v0.0.0
	github.com/Cooomma/ksm/ksm v0.0.0
	github.com/easonlin404/ksm v0.0.0-20190924154421-5e58f65580c0
	github.com/labstack/echo/v4 v4.1.11
)
