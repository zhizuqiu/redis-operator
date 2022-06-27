package sm4

import (
	"encoding/hex"
)

var Sm4Key = []byte("@*H7*Xb%jBiDq*Mb")

func EncryptSm4(data, key []byte) (string, error) {
	plainTextWithPadding := PKCS5Padding(data, BlockSize)
	cipherText, err := ECBEncrypt(key, plainTextWithPadding)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(cipherText), nil
}

func DecryptSm4(d, key []byte) (string, error) {
	data, err := hex.DecodeString(string(d))
	if err != nil {
		return "", err
	}
	plainTextWithPadding, err := ECBDecrypt(key, data)
	if err != nil {
		return "", err
	}
	plainText := PKCS5UnPadding(plainTextWithPadding)
	return string(plainText), nil
}
