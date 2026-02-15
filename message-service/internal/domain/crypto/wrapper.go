package crypto

type WrapperDEK interface {
	WrapDEK(dek, kek []byte) ([]byte, error)
	UnwrapDEK(dek, kek []byte) ([]byte, error)
}
