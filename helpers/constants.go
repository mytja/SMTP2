package helpers

var JwtSigningKey []byte = StringToBytearray("46ad2cb520028e1f5e2eab8d860a547353ddbabdb6affb923c075c92518c7e02")

const JwtIssuer = "SMTP2AuthCA"

const EnableDeletingOnUnknownError = true
