package main

import (
	"fmt"
	"os"
	"struct-go/src/internal/commands"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "struct_go",
		Short: "CLI para developers en Go",
		Long:  "CLI para crear estructuras base de proyectos APIs y microservicios en Go",
	}

	rootCmd.AddCommand(commands.ComponentCmd)
	rootCmd.AddCommand(commands.AddServiceCmd)

	rootCmd.AddCommand(commands.AddRestClientCmd)

	rootCmd.AddCommand(commands.AddGatewayCmd)
	rootCmd.AddCommand(commands.AddServiceGatewayCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
