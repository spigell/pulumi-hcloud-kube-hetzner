package distributions

const (
	K3SDistrName = "k3s"
	TalosDistrName = "talos"
)

type Distribution interface {
	Validate() error
}
