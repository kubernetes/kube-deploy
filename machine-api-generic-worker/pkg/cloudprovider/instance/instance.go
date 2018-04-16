package instance

// Instance represents a instance on the cloud provider
type Instance interface {
	Name() string
	ID() string
	Addresses() []string
}
