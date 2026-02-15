package crypto

type AESGCMEncryptor interface {
	Encode(plaintext, dek []byte) ([]byte, []byte, error)
	Decode(cipherText, nonce, dek []byte) ([]byte, error)
}
