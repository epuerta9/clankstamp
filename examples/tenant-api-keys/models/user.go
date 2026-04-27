package models

type User struct {
	ID       int
	Email    string
	TenantID int
}

// APIKey is a tenant-scoped credential for machine clients.
// The hash is bcrypt-encoded; the plaintext key is shown to the user once.
type APIKey struct {
	ID         int
	TenantID   int
	Hash       string
	LastUsedAt int64
	RevokedAt  int64
}
