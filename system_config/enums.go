package system_config

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type Mode string

const (
	Standalone Mode = "standalone"
	Cluster    Mode = "cluster"
)

type AMQPProtocol string

const (
	AMQP  AMQPProtocol = "amqp"
	AMQPS AMQPProtocol = "amqps"
)

type PubSubMode string

const (
	LocalPubSub  PubSubMode = "local"
	RemotePubSub PubSubMode = "remote"
)

type TaskQueueMode string

const (
	LocalTaskQueue  TaskQueueMode = "local"
	RemoteTaskQueue TaskQueueMode = "remote"
)
