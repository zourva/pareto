package broker

// MQ is the broker abstraction.
type MQ interface {
	// Startup starts the broker.
	Startup() error

	// Shutdown stops the broker.
	Shutdown() error
}

// MQClientOptions defines options provided to client when connecting.
type MQClientOptions struct {
}

// MQClient abstracts the broker client logic.
type MQClient interface {
	// Connect connects to the given broker using the provided conf.
	Connect(*MQClientOptions) error

	// Disconnect breaks connection with the broker.
	Disconnect()
}
