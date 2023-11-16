package distributions

type Distribution interface {
	Validate() error
}