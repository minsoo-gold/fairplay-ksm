module github.com/Coooomma/ksm/ksm

go 1.13

require (
	github.com/Cooomma/ksm/crypto v0.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
)

replace github.com/Cooomma/ksm/crypto v0.0.0 => ../crypto
