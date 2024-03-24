package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/service_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/worker"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
)

func init() {
	rootCmd.AddCommand(taskQueueCmd)
	taskQueueCmd.AddCommand(taskQueueListCmd)
	taskQueueCmd.AddCommand(taskQueueInspectCmd)
	taskQueueCmd.AddCommand(taskQueuePurgeCmd)
}

var taskQueueCmd = &cobra.Command{
	Use:   "tq",
	Short: "Task Queue Service",
	Long:  `Task Queue Service`,
	Run: func(cmd *cobra.Command, args []string) {
		// print help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var taskQueuePurgeCmd = &cobra.Command{
	Use:     "purge",
	Short:   "Purge Task Queue",
	Long:    `Purge messages of queues`,
	Example: `swiftwave tq purge <all|queue_name>`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printError("Queue name is required or use 'all' to purge all queues")
			printInfo(cmd.Example)
			return
		}
		dbClient, err := db.GetClient(config.LocalConfig, 2)
		if err != nil {
			printError("Failed to connect to database: " + err.Error())
			return
		}
		taskQueueClient, err := service_manager.FetchTaskQueueClient(config, dbClient)
		if err != nil {
			printError("Failed to fetch task queue client: " + err.Error())
			return
		}
		if args[0] == "all" {
			printInfo("Purging all queues ...")
			w := tabwriter.NewWriter(os.Stdout, 4, 1, 5, ' ', 0)
			fs := "%s\t%s\t%s\n"
			_, _ = fmt.Fprintf(w, fs, "QUEUE NAME", "STATUS", "ERROR")
			_, _ = fmt.Fprintf(w, fs, "----------", "------", "-----")
			isAllQueuesPurged := true
			for _, queue := range worker.Queues() {
				err = taskQueueClient.PurgeQueue(queue)
				if err != nil {
					isAllQueuesPurged = false
					_, _ = fmt.Fprintf(w, fs, queue, "FAILED", err.Error())
				} else {
					_, _ = fmt.Fprintf(w, fs, queue, "SUCCESS", "")
				}
			}
			_ = w.Flush()
			if isAllQueuesPurged {
				printSuccess("All queues purged successfully")
			} else {
				printError("Failed to purge all queues")
			}
		} else {
			if !isExistsInList(worker.Queues(), args[0]) {
				printError("Queue does not exist")
				return
			}
			printInfo("Purging queue ...")
			err = taskQueueClient.PurgeQueue(args[0])
			if err != nil {
				printError("Failed to purge queue: " + err.Error())
			} else {
				printSuccess("Queue purged successfully")
			}
		}
	},
}

var taskQueueListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all the queues names",
	Long:  `List all the queues names`,
	Run: func(cmd *cobra.Command, args []string) {
		queues := worker.Queues()
		// sort the queues
		sort.Strings(queues)
		w := tabwriter.NewWriter(os.Stdout, 4, 1, 1, ' ', 0)
		fs := "%s\t%s\n"
		_, _ = fmt.Fprintf(w, fs, "SR", "QUEUE NAME")
		_, _ = fmt.Fprintf(w, fs, "--", "----------")
		for i, queue := range queues {
			_, _ = fmt.Fprintf(w, fs, strconv.Itoa(i+1), queue)
		}
		_ = w.Flush()
	},
}

var taskQueueInspectCmd = &cobra.Command{
	Use:     "inspect",
	Short:   "Inspect a queue and print the messages",
	Long:    `Inspect a queue, will return total messages in the queue and the messages`,
	Example: `swiftwave tq inspect <queue_name>`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printError("Queue name is required")
			printInfo(cmd.Example)
			return
		}
		if !isExistsInList(worker.Queues(), args[0]) {
			printError("Queue does not exist")
			return
		}
		dbClient, err := db.GetClient(config.LocalConfig, 2)
		if err != nil {
			printError("Failed to connect to database: " + err.Error())
			return
		}
		taskQueueClient, err := service_manager.FetchTaskQueueClient(config, dbClient)
		if err != nil {
			printError("Failed to fetch task queue client: " + err.Error())
			return
		}
		printInfo("Inspecting queue ...")
		msg, err := taskQueueClient.ListMessages(args[0])
		if err != nil {
			printError("Failed to list messages: " + err.Error())
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 10, 1, 1, ' ', 0)
		fs := "%s\t%s\n"
		_, _ = fmt.Fprintf(w, fs, "NO", "MESSAGE")
		for i, msg := range msg {
			_, _ = fmt.Fprintf(w, fs, strconv.Itoa(i+1), msg)
		}
		_ = w.Flush()
		// if redis, print _processing queue's messages
		if config.SystemConfig.TaskQueueConfig.Mode == system_config.RemoteTaskQueue && config.SystemConfig.TaskQueueConfig.RemoteTaskQueueType == system_config.RedisQueue {
			fmt.Println("")
			printWarning("[Redis queue] printing on-the-fly messages list\n")
			processingMsg, err := taskQueueClient.ListMessages(fmt.Sprintf("%s_processing", args[0]))
			if err != nil {
				printError("Failed to list processing messages: " + err.Error())
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 10, 1, 1, ' ', 0)
			fs := "%s\t%s\n"
			_, _ = fmt.Fprintf(w, fs, "NO", "MESSAGE")
			for i, msg := range processingMsg {
				_, _ = fmt.Fprintf(w, fs, strconv.Itoa(i+1), msg)
			}
			_ = w.Flush()
		}
	},
}

func isExistsInList(list []string, item string) bool {
	for _, l := range list {
		if l == item {
			return true
		}
	}
	return false
}
