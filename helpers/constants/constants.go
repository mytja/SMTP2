package constants

import "github.com/mytja/SMTP2/helpers"

var JWT_SIGNING_KEY []byte = helpers.StringToBytearray("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")
var JWT_ISSUER = "SMTP2AuthCA"

// ENABLE_SMPT2_SSV SMPT2 SSV stands for SMTP2 Sender Server Verification
var ENABLE_SMPT2_SSV = true

var (
	SERVER_URL string
	DB_NAME    string
)
