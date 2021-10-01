package helpers

func BytearrayToString(t []byte) string {
	return string(t[:])
}

func StringToBytearray(t string) []byte {
	return []byte(t)
}
