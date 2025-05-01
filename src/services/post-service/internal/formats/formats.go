package formats

type Marshaler interface {
	Marshal() ([]byte, error)
}
