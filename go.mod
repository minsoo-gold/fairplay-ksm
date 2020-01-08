module github.com/Coooomma/ksm

go 1.13

require (
	github.com/Cooomma/ksm/crypto v0.0.0
	github.com/Cooomma/ksm/ksm v0.0.0
	github.com/labstack/echo/v4 v4.1.11
)

replace (
	github.com/Cooomma/ksm/crypto => ./crypto
	github.com/Cooomma/ksm/ksm => ./ksm
)
