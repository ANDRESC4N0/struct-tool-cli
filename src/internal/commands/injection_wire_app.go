package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"struct-go/src/internal/utils"
	"struct-go/src/internal/utils/logs"
)

func addInjectionApp(componentName string, injectionType string) {
	templateFilePath := utils.GetTemplateFilePath("app_template.txt")
	appPath := filepath.Join(apiDir, "app", "app.go")

	// Verificar si la ruta del app ya existe
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		logs.Echo("La ruta 'cmd/api/app' no existe.")
		logs.Chapter("Creando app/app.go")
		utils.CreateFile(appPath, templateFilePath, "")
	}

	// Buscar para insertar nueva inyección
	lines, err := utils.ReadFileLines(appPath)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al leer el archivo app.go: %s", err))
		return
	}

	namePascalCase := utils.ToPascalCase(componentName)
	targetLine := fmt.Sprintf("var %sRouterSet = wire.NewSet(", injectionType)
	found := false
	linePosition := 0

	for i, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(targetLine) {
			found = true
			injectionLine := getInjectionLine(componentName, namePascalCase, injectionType)

			lines = append(lines[:i+1], append([]string{injectionLine}, lines[i+1:]...)...)
			break
		}

		if !found && strings.TrimSpace(line) == ")" {
			linePosition = i
		}
		if !found && strings.TrimSpace(line) == "panic(wire.Build(" {
			lines = append(lines[:i+1], append([]string{fmt.Sprintf("\t\t%sRouterSet,\n", injectionType)}, lines[i+1:]...)...)
		}
	}

	if !found && linePosition > 0 {
		injectionLine := getInjectionLine(componentName, namePascalCase, injectionType)
		lines = append(lines[:linePosition+1], append([]string{fmt.Sprintf("\nvar %sRouterSet = wire.NewSet(", injectionType)}, lines[linePosition+1:]...)...)
		lines = append(lines[:linePosition+2], append([]string{injectionLine}, lines[linePosition+2:]...)...)
		lines = append(lines[:linePosition+3], append([]string{")"}, lines[linePosition+3:]...)...)
	}

	err = utils.WriteFileLines(appPath, lines)
	if err != nil {
		logs.Error(fmt.Sprintf("Error al escribir en el archivo app.go: %s", err))
	}
}

func getInjectionLine(componentName string, namePascalCase string, injectionType string) string {
	switch injectionType {
	case "components":
		return fmt.Sprintf("\tproviders.Init%sService, wire.Bind(new(%s.I%sService), new(*%s.%sService)),", namePascalCase, componentName, namePascalCase, componentName, namePascalCase)
	case "controllers":
		return fmt.Sprintf("\tproviders.Init%sController, wire.Bind(new(%s.I%sController), new(*%s.%sController)),", namePascalCase, componentName, namePascalCase, componentName, namePascalCase)
	case "repositories":
		return fmt.Sprintf("\tproviders.Init%sRepository, wire.Bind(new(%s.I%sRepository), new(*%s.%sRepository)),", namePascalCase, componentName, namePascalCase, componentName, namePascalCase)
	case "clients":
		return fmt.Sprintf("\tproviders.Init%sClient, wire.Bind(new(clients.I%sClient), new(*clients.%sClient)),", namePascalCase, componentName, namePascalCase)
	case "gateway":
		return fmt.Sprintf("\tproviders.Init%sGateway, wire.Bind(new(gateway.I%sGateway), new(*gateway.%sGateway)),", namePascalCase, componentName, namePascalCase)
	case "middlewares":
		return fmt.Sprintf("\tproviders.Init%sMiddleware, wire.Bind(new(middlewares.I%sMiddleware), new(*middlewares.%sMiddleware)),", namePascalCase, componentName, namePascalCase)
	default:
		panic("Tipo de inyección no válido")
	}
}
