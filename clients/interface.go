package clients

// Interface is client interface
type Interface interface {
	Args() []string
	Command() []string
	HomeDir() string
	Image() string
}
