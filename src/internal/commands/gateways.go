package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"struct-go/src/internal/utils"
	"struct-go/src/internal/utils/logs"

	"github.com/spf13/cobra"
)

var gatewayName string
var rest string

var AddGatewayCmd = &cobra.Command{
	Use:   "add_gateway",
	Short: "Agrega un gateway a la carpeta 'gateway'.",
	Run:   createGateway,
}

var AddServiceGatewayCmd = &cobra.Command{
	Use:   "add_service_gateway",
	Short: "Agrega un servicio al gateway.",
	Run:   addServiceGateway,
}

func init() {
	AddGatewayCmd.Flags().StringVarP(&gatewayName, "gateway_name", "g", "", "El nombre del gateway a crear.")
	AddGatewayCmd.Flags().StringVarP(&rest, "rest", "r", "", "El método rest a agregar al gateway.")
	AddServiceGatewayCmd.Flags().StringVarP(&gatewayName, "gateway_name", "g", "", "El nombre del gateway a crear.")
	AddServiceGatewayCmd.Flags().StringVarP(&rest, "rest", "r", "", "El método rest a agregar al gateway.")
	// ComponentCmd.MarkFlagRequired("component_name")
}

func createGateway(cmd *cobra.Command, args []string) {
	name := utils.Prompt("Ingresa el nombre del gateway: ")
	gatewayDir := filepath.Join(apiDir, "gateway")

	// Verificar si la ruta 'cmd/api/gateway' existe
	if _, err := os.Stat(gatewayDir); os.IsNotExist(err) || !utils.IsDir(gatewayDir) {
		logs.Echo("La ruta 'cmd/api/gateway' no existe.")
		logs.Chapter("Creando la ruta 'cmd/api/gateway'...")
		os.MkdirAll(gatewayDir, os.ModePerm)
	}

	// Verificar si el gateway ya existe
	gatewayFile := filepath.Join(gatewayDir, name+".go")
	if _, err := os.Stat(gatewayFile); !os.IsNotExist(err) {
		logs.Error(fmt.Sprintf("El gateway '%s' ya existe.", name))
		os.Exit(1)
	}

	logs.Echo(fmt.Sprintf("Creando el gateway %s...\n", gatewayFile))
	gatewayTemplateFilePath := utils.GetTemplateFilePath("gateway_template.txt")
	utils.CreateFile(gatewayFile, gatewayTemplateFilePath, name)

	// Agregar método
	gatewayName = name
	addServiceGateway(cmd, args)
	logs.Chapter(fmt.Sprintf("Gateway %s creado exitosamente.", name))
}

func addServiceGateway(cmd *cobra.Command, args []string) {
	rest, _ := cmd.Flags().GetString("rest")

	if gatewayName == "" {
		gatewayName = utils.Prompt("Ingresa el nombre del gateway: ")
	}

	if rest == "" {
		rest = utils.Prompt("Ingresa el método rest (GET, POST, PUT, DELETE): ")
	}

	var restTemplateFilePath string

	switch rest {
	case "GET":
		restTemplateFilePath = utils.GetTemplateFilePath("get_template.txt")
	case "POST":
		restTemplateFilePath = utils.GetTemplateFilePath("post_template.txt")
	case "PUT":
		restTemplateFilePath = utils.GetTemplateFilePath("put_template.txt")
	case "DELETE":
		restTemplateFilePath = utils.GetTemplateFilePath("delete_template.txt")
	default:
		logs.Error(fmt.Sprintf("Opción rest->%s no válida.", rest))
		os.Exit(1)
	}

	gatewayFilePath := filepath.Join(apiDir, "gateway", gatewayName+".go")
	if !utils.FileExists(gatewayFilePath) {
		logs.Error(fmt.Sprintf("El gateway '%s' no existe.", gatewayName))
		os.Exit(1)
	}

	addRest(gatewayFilePath, restTemplateFilePath)
}

func addRest(gatewayFilePath, restTemplateFilePath string) {
	gatewayName = utils.ToPascalCase(gatewayName)
	logs.Echo(fmt.Sprintf("Agregando método al gateway %s...", gatewayFilePath))

	lines, err := utils.ReadFileLines(gatewayFilePath)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo del gateway: %s", err.Error()))
		os.Exit(1)
	}

	// Buscar interface para agregar la función
	found := false
	nameFunction := utils.Prompt("Ingrese el nombre de la función")
	modelRequest := utils.Prompt("Ingrese el nombre del modelo de petición")
	modelResponse := utils.Prompt("Ingrese el nombre del modelo de respuesta")
	for i, line := range lines {
		if strings.Contains(line, fmt.Sprintf("type I%sGateway interface {", gatewayName)) {
			found = true
			functionLine := fmt.Sprintf("\t%s(search requests.%s, ctx context.Context, txn *newrelic.Transaction) (*responses.%s, error)", nameFunction, modelRequest, modelResponse)
			lines = append(lines[:i+1], append([]string{functionLine}, lines[i+1:]...)...)
		}
	}

	if !found {
		logs.Error(fmt.Sprintf("No se encontró la interface I%sGateway en el archivo %s", gatewayName, gatewayFilePath))
		os.Exit(1)
	}

	// Añadir la función al archivo
	templateContent, err := ioutil.ReadFile(restTemplateFilePath)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo de la función: %s", err.Error()))
		os.Exit(1)
	}

	content := string(templateContent)
	content = strings.ReplaceAll(content, "{NameGateway}", gatewayName)
	content = strings.ReplaceAll(content, "{NameFunction}", nameFunction)
	content = strings.ReplaceAll(content, "{NameModelRequest}", modelRequest)
	content = strings.ReplaceAll(content, "{NameModelResponse}", modelResponse)
	lines = append(lines, "\n", content)

	// Escribir en el archivo
	err = utils.WriteFileLines(gatewayFilePath, lines)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al escribir en el archivo del gateway: %s", err.Error()))
		os.Exit(1)
	}
}
