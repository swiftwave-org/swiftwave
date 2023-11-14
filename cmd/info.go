package cmd

import (
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/system_config"
	"strconv"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print info of swiftwave",
	Long:  `Print info of swiftwave`,
	Run: func(cmd *cobra.Command, args []string) {
		printInfo("Version : " + systemConfig.Version)
		printInfo("Deployed in " + string(systemConfig.Mode) + " mode")
		printInfo("Domain pointed to current server > " + systemConfig.ServiceConfig.AddressOfCurrentNode)
		printInfo("Listening on " + systemConfig.ServiceConfig.BindAddress + ":" + strconv.Itoa(systemConfig.ServiceConfig.BindPort))
		printInfo("Service accessible at https://" + systemConfig.ServiceConfig.AddressOfCurrentNode + ":" + strconv.Itoa(systemConfig.ServiceConfig.BindPort))
		println("---------------------------------------------------")
		printInfo("Docker UNIX socket : " + systemConfig.ServiceConfig.DockerUnixSocketPath)
		printInfo("Overlay network : " + systemConfig.ServiceConfig.NetworkName)
		ports := ""
		for _, port := range systemConfig.ServiceConfig.RestrictedPorts {
			ports += strconv.Itoa(port) + ", "
		}
		printInfo("Data will be stored in : " + systemConfig.ServiceConfig.DataDir)
		println("---------------------------------------------------")
		printInfo("Let's Encrypt E-mail : " + systemConfig.LetsEncryptConfig.EmailID)
		printInfo("Let's Encrypt Account Key : " + systemConfig.LetsEncryptConfig.AccountPrivateKeyPath)
		println("---------------------------------------------------")
		printInfo("Restricted Ports : " + ports)
		printInfo("HAProxy Service name : " + systemConfig.HAProxyConfig.ServiceName)
		printInfo("HAProxy Image Used : " + systemConfig.HAProxyConfig.DockerImage)
		printInfo("HAProxy Dataplane API Unix Socket : " + systemConfig.HAProxyConfig.UnixSocketPath)
		printInfo("HAProxy Admin & Password : " + systemConfig.HAProxyConfig.User + ":" + systemConfig.HAProxyConfig.Password)
		println("---------------------------------------------------")
		printInfo("Postgres Address : " + systemConfig.PostgresqlConfig.Host + ":" + strconv.Itoa(systemConfig.PostgresqlConfig.Port))
		printInfo("Postgres User : " + systemConfig.PostgresqlConfig.User)
		printInfo("Postgres Password : " + systemConfig.PostgresqlConfig.Password)
		printInfo("Postgres Database : " + systemConfig.PostgresqlConfig.Database)
		printInfo("Postgres Timezone : " + systemConfig.PostgresqlConfig.TimeZone)
		msg := "Disabled"
		if systemConfig.ServiceConfig.AutoMigrateDatabase {
			msg = "Enabled"
		}
		printInfo("Database Auto Migration : " + msg)
		println("---------------------------------------------------")
		printInfo("Pubsub mode : " + string(systemConfig.PubSubConfig.Mode))
		printInfo("Pubsub Buffer Size : " + strconv.Itoa(systemConfig.PubSubConfig.BufferLength))
		if systemConfig.PubSubConfig.Mode == system_config.RemotePubSub {
			printInfo("Redis Address : " + systemConfig.PubSubConfig.RedisConfig.Host + ":" + strconv.Itoa(systemConfig.PubSubConfig.RedisConfig.Port))
			printInfo("Redis Password : " + systemConfig.PubSubConfig.RedisConfig.Password)
			printInfo("Redis Database : " + strconv.Itoa(systemConfig.PubSubConfig.RedisConfig.DatabaseID))
		}
		println("---------------------------------------------------")
		printInfo("Task Queue mode : " + string(systemConfig.TaskQueueConfig.Mode))
		printInfo("Max Outstanding Messages Per Queue : " + strconv.Itoa(systemConfig.TaskQueueConfig.MaxOutstandingMessagesPerQueue))
		if systemConfig.TaskQueueConfig.Mode == system_config.RemoteTaskQueue {
			printInfo("AMQP Address : " + string(systemConfig.TaskQueueConfig.AMQPConfig.Protocol) + "://" + systemConfig.TaskQueueConfig.AMQPConfig.Host)
			printInfo("AMQP User : " + systemConfig.TaskQueueConfig.AMQPConfig.User)
			printInfo("AMQP Password : " + systemConfig.TaskQueueConfig.AMQPConfig.Password)
			printInfo("AMQP Virtual Host : " + systemConfig.TaskQueueConfig.AMQPConfig.VHost)
		}
	},
}
