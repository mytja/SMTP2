package constants

import "github.com/mytja/SMTP2/helpers"

var JwtSigningKey []byte = helpers.StringToBytearray("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")
var JwtIssuer = "SMTP2AuthCA"

// ENABLE_SMTP2_SSV SMPT2 SSV stands for SMTP2 Sender Server Verification
var EnableSmtp2Ssv = true

var ForceHttps = false
var ForceHttpsForMailDomain = false

var (
	ServerUrl string
	DbName    string
)
