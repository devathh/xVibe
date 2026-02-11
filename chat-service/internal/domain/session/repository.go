package session

type JWTManager interface {
	Validate(tokenString string) (*Claims, error)
}
