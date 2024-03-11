package cmd

import (
	"github.com/fatih/color"
	"net"
	"os"
	"os/exec"
)

func checkIfFolderExists(folder string) bool {
	cmd := exec.Command("ls", folder)
	err := cmd.Run()
	return err == nil
}

func createFolder(folder string) error {
	return os.MkdirAll(folder, 0711)
}

func checkIfFileExists(file string) bool {
	cmd := exec.Command("ls", file)
	err := cmd.Run()
	return err == nil
}

func checkIfPortIsInUse(port string) bool {
	// Attempt to establish a connection to the address
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			printError("Error closing connection")
		}
	}(conn)
	return true
}

func openFileInEditor(filePath string) {
	// Check if the $EDITOR environment variable is set
	editor := os.Getenv("EDITOR")

	if editor != "" {
		// $EDITOR is set, use it to open the file
		cmd := exec.Command(editor, filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			printError("Error opening file with " + editor)
		}
	} else {
		// $EDITOR is not set, try using mimeopen
		cmd := exec.Command("mimeopen", "-d", filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			printError("Error opening file with mimeopen")
			printError("Set the $EDITOR environment variable to open the file with your preferred editor")
		}
	}
}

func printSuccess(message string) {
	color.Green(TickSymbol + " " + message)
}

func printError(message string) {
	color.Red(CrossSymbol + " " + message)
}

func printInfo(message string) {
	color.Blue(InfoSymbol + " " + message)
}