package shared

// Client is a protocol client
type Client interface {
	Args() []string
	Command() []string
	HomeDir() string
	Image() string
}
