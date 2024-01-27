package yeahapi

type FlashKind int

const (
	ErrFlashKind = FlashKind(iota + 1)
	InfoFlashKind
	SuccessFlashKind
)

type Flash struct {
	Kind    FlashKind
	Message string
}
