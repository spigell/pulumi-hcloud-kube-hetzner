package distributions

const (
	K3SDistrName = "k3s"
)

type Distribution interface {
	Validate() error
}
