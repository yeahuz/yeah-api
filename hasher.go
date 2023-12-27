package yeahapi

type ArgonParams struct {
	SaltLen uint32
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

type ArgonHasher interface {
	Hash(b []byte) (string, error)
	Verify(s, encoded string) error
	Decode(encoded string) (p *ArgonParams, salt, hash []byte, err error)
}

type HighwayHasher interface {
	Hash(b []byte) (string, error)
}
