package commands

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"struct-go/src/internal/utils"

	"struct-go/src/internal/utils/logs"

	"github.com/spf13/cobra"
)

var componentName string
var serviceName string

var ComponentCmd = &cobra.Command{
	Use:   "create_component",
	Short: "Crea un componente en la ruta 'cmd/api'.",
	Run:   createComponent,
}

var AddServiceCmd = &cobra.Command{
	Use:   "add_service_component",
	Short: "Agrega un servicio a un componente.",
	Run:   addService,
}

func init() {
	ComponentCmd.Flags().StringVarP(&componentName, "component_name", "c", "", "El nombre del componente a crear.")
	AddServiceCmd.Flags().StringVarP(&componentName, "component_name", "c", "", "El nombre del componente al que se le agregará el servicio.")
	AddServiceCmd.Flags().StringVarP(&serviceName, "service_name", "s", "", "El nombre del servicio a agregar al componente.")
}

func createComponent(cmd *cobra.Command, args []string) {
	// Si el nombre del componente no se ha proporcionado como bandera,
	// solicitarlo al usuario
	if componentName == "" {
		reader := bufio.NewReader(os.Stdin)
		logs.Echo("Por favor, ingresa el nombre del componente: ")
		input, _ := reader.ReadString('\n')
		componentName = strings.TrimSpace(input)
	}

	// Construir la ruta del directorio del componente dentro de 'cmd/api'
	apiComponentsDir := filepath.Join(apiDir, "components")
	componentDir := filepath.Join(apiComponentsDir, componentName)

	// Verificar si la ruta 'cmd/api' existe
	if _, err := os.Stat(apiComponentsDir); os.IsNotExist(err) {
		logs.Error("La ruta 'cmd/api' no existe.")
		return
	}

	// Verificar si la ruta del componente ya existe
	if _, err := os.Stat(componentDir); !os.IsNotExist(err) {
		logs.Error(fmt.Sprintf("El componente '%s' ya existe en la ruta 'cmd/api'.", componentName))
		return
	}

	// Crear el directorio del componente dentro de 'cmd/api'
	err := os.MkdirAll(componentDir, os.ModePerm)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al crear el directorio del componente: %s", err))
		return
	}

	serviceFilePath := filepath.Join(componentDir, "service.go")
	controllerFilePath := filepath.Join(componentDir, "controller.go")

	serviceTemplateFilePath := utils.GetTemplateFilePath("service_template.txt")
	controllerTemplateFilePath := utils.GetTemplateFilePath("controller_template.txt")

	utils.CreateFile(serviceFilePath, serviceTemplateFilePath, componentName)
	utils.CreateFile(controllerFilePath, controllerTemplateFilePath, componentName)

	logs.Chapter(fmt.Sprintf("Componente '%s' creado exitosamente.", componentName))

	// Obtener las interfaces de servicios y gateways
	serviceInterfaces := findInterfaces(apiComponentsDir, `I(\w+)Service interface`)
	gatewayInterfaces := findInterfaces(filepath.Join(apiDir, "gateway"), `I(\w+)Gateway interface`)

	// Inyectar los servicios en el archivo service.go
	injections := injectionServices(apiComponentsDir, componentName, serviceInterfaces, gatewayInterfaces)
	providerFilePath := filepath.Join(apiDir, "app/providers/components.go")
	providerTemplateFilePath := utils.GetTemplateFilePath("provider_components_template.txt")

	if _, err := os.Stat(providerFilePath); os.IsNotExist(err) {
		utils.CreateFile(providerFilePath, providerTemplateFilePath, componentName)
	}

	file, err := os.OpenFile(providerFilePath, os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al abrir el archivo de proveedores: %s", err))
		return
	}
	defer file.Close()

	namePascal := utils.ToPascalCase(componentName)
	file.WriteString(fmt.Sprintf("\n\nfunc Init%sService(\n", namePascal))
	logs.Echo(fmt.Sprintf("Inyecciones para el componente '%s':", componentName))
	for key, value := range injections {
		camel := utils.ToCamelCase(key)
		file.WriteString(fmt.Sprintf("\t%s *%s,\n", camel, value))
	}
	file.WriteString(fmt.Sprintf(") *%s.%sService {\n", componentName, namePascal))
	file.WriteString(fmt.Sprintf("\treturn &%s.%sService{\n", componentName, namePascal))
	for key := range injections {
		camel := utils.ToCamelCase(key)
		file.WriteString(fmt.Sprintf("\t\t%s: %s,\n", key, camel))
	}
	file.WriteString("\t}\n}\n\n")

	// TODO: Inyectar el componente en el archivo 'app/app.go' y ejecutar 'go generate ./...'
	addInjectionApp(componentName, "components")
}

func addService(cmd *cobra.Command, args []string) {
	if componentName == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Ingrese el nombre del componente: ")
		componentName, _ = reader.ReadString('\n')
		componentName = strings.TrimSpace(componentName)
	}

	if serviceName == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Ingrese el nombre del servicio: ")
		serviceName, _ = reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)
	}

	apiComponentsDir := filepath.Join(apiDir, "components")
	serviceFilePath := filepath.Join(apiComponentsDir, componentName, "service.go")

	if !utils.FileExists(serviceFilePath) {
		logs.Error(fmt.Sprintf("Error: El archivo '%s' no existe en la ruta actual.", serviceFilePath))
		os.Exit(1)
	}

	file, err := os.OpenFile(serviceFilePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		logs.Error(fmt.Sprintf("Error: No se puede abrir el archivo '%s'.", serviceFilePath))
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	pascal := utils.ToPascalCase(componentName)
	targetLine := fmt.Sprintf("type I%sService interface {", pascal)
	inserted := false

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		if strings.Contains(line, targetLine) {
			lines = append(lines, fmt.Sprintf("\t%s(requestSearch interface{}, ctx context.Context, txn *newrelic.Transaction) (interface{}, error)\n", serviceName))
			inserted = true
		}
	}

	if err := scanner.Err(); err != nil {
		logs.Error(fmt.Sprintf("Error: No se puede leer el archivo '%s'.", serviceFilePath))
		os.Exit(1)
	}

	if !inserted {
		logs.Error(fmt.Sprintf("Error: No se encontró la línea '%s' en el archivo '%s'.", targetLine, serviceFilePath))
		os.Exit(1)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("func (service *%sService) %s(requestSearch interface{}, ctx context.Context, txn *newrelic.Transaction) (interface{}, error) {\n", pascal, serviceName))
	lines = append(lines, "\treturn nil, nil\n")
	lines = append(lines, "}")

	if err := file.Truncate(0); err != nil {
		logs.Error(fmt.Sprintf("Error: No se puede truncar el archivo '%s'.", serviceFilePath))
		os.Exit(1)
	}

	if _, err := file.Seek(0, 0); err != nil {
		logs.Error(fmt.Sprintf("Error: No se puede mover al principio del archivo '%s'.", serviceFilePath))
		os.Exit(1)
	}

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			logs.Error(fmt.Sprintf("Error: No se puede escribir en el archivo '%s'.", serviceFilePath))
			os.Exit(1)
		}
	}

	if err := writer.Flush(); err != nil {
		logs.Error(fmt.Sprintf("Error: No se puede guardar los cambios en el archivo '%s'.", serviceFilePath))
		os.Exit(1)
	}

	logs.Chapter(fmt.Sprintf("El contenido ha sido agregado exitosamente en el archivo '%s'.", serviceFilePath))
}

func findInterfaces(directory string, pattern string) map[string]map[string]interface{} {
	if _, err := os.Stat(directory); os.IsNotExist(err) || !utils.IsDir(directory) {
		logs.Echo(fmt.Sprintf("El directorio '%s' no existe.", directory))
		return make(map[string]map[string]interface{})
	}

	interfaces := make(map[string]map[string]interface{})

	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logs.Error(fmt.Sprintf("Error al acceder a '%s': %s", path, err))
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				logs.Error(fmt.Sprintf("Error al leer el archivo '%s': %s", path, err))
				return nil
			}

			matches := regexp.MustCompile(pattern).FindAllStringSubmatch(string(content), -1)
			if len(matches) > 0 {
				packageName := filepath.Base(filepath.Dir(path))
				for _, match := range matches {
					interfaceName := fmt.Sprintf("I%s", match[1])
					interfaces[interfaceName] = map[string]interface{}{
						"package": packageName,
					}
				}
			}
		}

		return nil
	})

	return interfaces
}

func injectionServices(apiComponentsDir string, componentName string, serviceInterfaces map[string]map[string]interface{}, gatewayInterfaces map[string]map[string]interface{}) map[string]string {
	result := make(map[string]string)
	listServices := make([]string, 0)
	msgPromt := "Ingrese los servicios que requieres en el componente (separados por espacios)"

	if len(serviceInterfaces) > 0 {
		logs.Echo("Se encontraron los siguientes nombres de interfaces de servicio:")
		for serviceName := range serviceInterfaces {
			logs.Echo(serviceName)
		}
		listCompService := utils.Prompt(msgPromt)
		listCompService = strings.TrimSpace(listCompService)
		if listCompService != "" {
			listServices = append(listServices, strings.Split(listCompService, " ")...)
		}
	} else {
		logs.Echo("No se encontraron nombres de interfaces de servicio.")
	}

	if len(gatewayInterfaces) > 0 {
		logs.Echo("Se encontraron los siguientes nombres de interfaces de gateways:")
		for gatewayName := range gatewayInterfaces {
			logs.Echo(gatewayName)
		}
		listGateway := utils.Prompt(msgPromt)
		listGateway = strings.TrimSpace(listGateway)
		if listGateway != "" {
			listServices = append(listServices, strings.Split(listGateway, " ")...)
		}
	} else {
		logs.Echo("No se encontraron nombres de interfaces de gateway.")
	}

	// Agregar los servicios al archivo service.go
	serviceFilePath := filepath.Join(apiComponentsDir, componentName, "service.go")
	content, err := ioutil.ReadFile(serviceFilePath)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo '%s': %s", serviceFilePath, err))
		os.Exit(1)
	}

	newContent := string(content)

	// Buscar la línea donde se debe agregar el nuevo contenido
	targetLine := fmt.Sprintf("type %sService struct {", utils.ToPascalCase(componentName))
	lineIndex := strings.Index(newContent, targetLine)
	if lineIndex == -1 {
		logs.Error(fmt.Sprintf("Error: No se encontró la línea '%s' en el archivo '%s'.", targetLine, serviceFilePath))
		os.Exit(1)
	}

	// Insertar el nuevo contenido después de la línea objetivo
	for _, serviceName := range listServices {
		packageService := ""
		if serviceInfo, ok := serviceInterfaces[serviceName]; ok {
			packageService = fmt.Sprintf("%s.%sService", serviceInfo["package"], serviceName)
		} else if gatewayInfo, ok := gatewayInterfaces[serviceName]; ok {
			packageService = fmt.Sprintf("%s.%sGateway", gatewayInfo["package"], serviceName)
		}

		if packageService != "" {
			result[serviceName[1:]] = packageService
			insertContent := fmt.Sprintf("\n\t%s %s", serviceName[1:], packageService)
			newContent = newContent[:lineIndex+len(targetLine)] + insertContent + newContent[lineIndex+len(targetLine):]
		}
	}

	// Escribir el contenido modificado en el archivo
	err = ioutil.WriteFile(serviceFilePath, []byte(newContent), 0644)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al escribir en el archivo '%s': %s", serviceFilePath, err))
		os.Exit(1)
	}

	logs.Chapter(fmt.Sprintf("Se agregaron los servicios al archivo '%s'.", serviceFilePath))

	return result
}
