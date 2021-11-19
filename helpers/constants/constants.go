package constants

import "github.com/mytja/SMTP2/helpers"

var JwtSigningKey []byte = helpers.StringToBytearray("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")

const JwtIssuer = "SMTP2AuthCA"

// EnableSmtp2Ssv ENABLE_SMTP2_SSV SMPT2 SSV stands for SMTP2 Sender Server Verification
const EnableSmtp2Ssv = true

const ForceHttps = false
const ForceHttpsForMailDomain = false

const EnableDeletingOnUnknownError = true

var (
	ServerUrl string
	DbName    string
)
