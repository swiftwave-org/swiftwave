package main

import (
	"fmt"
	"os"
	// "sync"
	. "keroku/m/haproxy_manager"
)



func main() {
	// var wg sync.WaitGroup

	// Create a new HAProxySocket
	var haproxySocket = HAProxySocket{};
	haproxySocket.InitTcpSocket("localhost", 5555);
	haproxySocket.Auth("admin", "mypassword");
	errFound := false;
	transaction_id, err := haproxySocket.FetchNewTransactionId()
	if err != nil {
		print("Error while fetching HAProxy version: " + err.Error())
		os.Exit(1)
		return
	}

	// Add backend
	// if err != nil {
	// 	errFound = true;
	// }else{
	// 	err := haproxySocket.AddBackend(transaction_id, "minc-service", 3000, 3);
	// 	if err != nil {
	// 		errFound = true;
	// 	}
	// 	fmt.Println("Add backend")
	// }

	// Update backend
	// if err != nil {
	// 	errFound = true;
	// }else{
	// 	err := haproxySocket.UpdateBackend(transaction_id, "minc-service-new3", 5000, 1, "minc-service", 3000, 2);
	// 	if err != nil {
	// 		errFound = true;
	// 	}
	// 	fmt.Println(err)
	// 	fmt.Println("Update backend")
	// }

	if errFound {
		fmt.Println("Deleting transaction: "+transaction_id)
		haproxySocket.DeleteTransaction(transaction_id)
		fmt.Println("Error found")
	}else{
		fmt.Println("Committing transaction: "+transaction_id)
		haproxySocket.CommitTransaction(transaction_id)
		fmt.Println("No error found")
	}


	// Wait for events
	// wg.Wait()
	fmt.Println("done")
}
