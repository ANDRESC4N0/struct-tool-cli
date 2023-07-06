package commands

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"struct-go/src/internal/utils"
	"struct-go/src/internal/utils/logs"

	"github.com/spf13/cobra"
)

var AddRestClientCmd = &cobra.Command{
	Use:   "add_rest_client",
	Short: "Agrega un cliente rest a la carpeta 'clients'.",
	Run:   createRestClient,
}

func init() {
}

func createRestClient(cmd *cobra.Command, args []string) {
	logs.Echo("Creando cliente rest_client")

	templateFileInterface := utils.GetTemplateFilePath("rest_client_interface_template.txt")
	createInterfaceClient(templateFileInterface, "IRestClient")

	templateFileClient := utils.GetTemplateFilePath("rest_client_template.txt")
	createClientImpl(templateFileClient, "rest_client.go")

	templateModelFile := utils.GetTemplateFilePath("rest_client_model_template.txt")
	createModel(templateModelFile, "restclient/rest_client.go")
}

func createInterfaceClient(templateFilePath, interfaceName string) {
	interfacesFile := filepath.Join(apiDir, "clients", "interfaces.go")
	if !utils.FileExists(interfacesFile) {
		logs.Echo(fmt.Sprintf("El archivo interfaces.go no existe. Creando el archivo con la interfaz %s...", interfaceName))

		utils.CreateFile(interfacesFile, templateFilePath, "")

		lines, err := utils.ReadFileLines(interfacesFile)
		if err != nil {
			logs.Error(fmt.Sprintf("Error al leer el archivo de: %s", interfacesFile))
			return
		}

		lines = append(lines[:0], append([]string{"package clients\n\n"}, lines[0:]...)...)

		if err := utils.WriteFileLines(interfacesFile, lines); err != nil {
			logs.Error(fmt.Sprintf("Error al escribir el archivo de: %s", interfacesFile))
		}

		return
	}

	lines, err := utils.ReadFileLines(interfacesFile)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo de: %s", interfacesFile))
		return
	}

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("type %s interface {", interfaceName)) {
			logs.Chapter(fmt.Sprintf("La interfaz %s ya está definida en el archivo interfaces.go.", interfaceName))
			return
		}
	}

	logs.Echo(fmt.Sprintf("La interfaz %s no está definida en el archivo interfaces.go. Agregando la interfaz...", interfaceName))
	templateFilePath = utils.GetTemplateFilePath(templateFilePath)

	content, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo de plantilla: %s", err))
		return
	}

	lines = append(lines[:len(lines)-1], append([]string{string(content)}, lines[len(lines)-1:]...)...)

	if err := utils.WriteFileLines(interfacesFile, lines); err != nil {
		logs.Error(fmt.Sprintf("Error al escribir el archivo de: %s", interfacesFile))
	}
}

func createClientImpl(templateFilePath, clientName string) {
	restClientFile := filepath.Join(apiDir, "clients", clientName)
	if utils.FileExists(restClientFile) {
		restClientContent, err := ioutil.ReadFile(restClientFile)
		if err != nil {
			logs.Error(fmt.Sprintf("Error al leer el archivo %s: %s", clientName, err))
			return
		}

		templateContent, err := ioutil.ReadFile(templateFilePath)
		if err != nil {
			logs.Error(fmt.Sprintf("Error al leer el archivo de plantilla: %s", err))
			return
		}

		if strings.TrimSpace(string(templateContent)) == strings.TrimSpace(string(restClientContent)) {
			logs.Echo(fmt.Sprintf("El archivo %s ya existe y tiene el contenido correcto.", clientName))
		} else {
			logs.Echo(fmt.Sprintf("El archivo %s existe pero tiene un contenido distinto al esperado. Actualizando el archivo...", clientName))

			err = ioutil.WriteFile(restClientFile, templateContent, 0644)
			if err != nil {
				logs.Error(fmt.Sprintf("Error al actualizar el archivo %s: %s", clientName, err))
			}
		}
	} else {
		logs.Echo(fmt.Sprintf("El archivo %s no existe. Creando el archivo con el contenido del template...", clientName))
		utils.CreateFile(restClientFile, templateFilePath, "")
	}
}

// TODO: Corregir
func createModel(templateFilePath, modelPath string) {
	modelFile := filepath.Join(apiDir, "models", modelPath)
	if !utils.FileExists(modelFile) {
		logs.Echo(fmt.Sprintf("El modelo %s no existe. Creando el archivo con el contenido del template...", modelPath))
		utils.CreateFile(modelFile, templateFilePath, "")
		return
	}

	restClientContent, err := ioutil.ReadFile(modelFile)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo %s: %s", modelPath, err))
		return
	}

	templateContent, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo de plantilla: %s", err))
		return
	}

	if strings.TrimSpace(string(templateContent)) == strings.TrimSpace(string(restClientContent)) {
		logs.Chapter("El modelo ya existe y tiene el contenido correcto.")
		return
	}

	logs.Echo("El modelo existe pero tiene un contenido distinto al esperado. Actualizando el archivo...")

	err = ioutil.WriteFile(modelFile, templateContent, 0644)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al actualizar el archivo %s: %s", modelPath, err))
	}
}
