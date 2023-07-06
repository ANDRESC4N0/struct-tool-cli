package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"struct-go/internal/utils/logs"

	"github.com/manifoldco/promptui"
)

func ToCamelCase(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, " ", "_"))
	words := strings.Split(text, "_")
	result := words[0] + strings.Join(words[1:], "")
	return strings.ToLower(result[:1]) + result[1:]
}

func ToPascalCase(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, " ", "_"))
	words := strings.Split(text, "_")
	result := ""
	for _, word := range words {
		result += strings.Title(word)
	}
	return result
}

func ToSnakeCase(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, " ", "_"))
	result := ""
	for i, c := range text {
		if i > 0 && c >= 'A' && c <= 'Z' && text[i-1] != '_' {
			result += "_"
		}
		result += string(c)
	}
	return strings.ToLower(result)
}

func CreateFile(filePath, templateFilePath, componentName string) {
	namePascalCase := ToPascalCase(componentName)

	templateFile, err := os.Open(templateFilePath)
	if err != nil {
		logs.Error(fmt.Sprintf("No se puede abrir el archivo de plantilla '%s'", templateFilePath))
		os.Exit(1)
	}
	defer templateFile.Close()

	templateContent := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := templateFile.Read(buffer)
		if n > 0 {
			templateContent = append(templateContent, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	content := string(templateContent)
	content = strings.ReplaceAll(content, "{NamePackage}", componentName)
	content = strings.ReplaceAll(content, "{NameComponent}", namePascalCase)
	content = strings.ReplaceAll(content, "{NameGateway}", namePascalCase)

	file, err := os.Create(filePath)
	if err != nil {
		logs.Error(fmt.Sprintf("No se puede crear el archivo '%s'", filePath))
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		logs.Error(fmt.Sprintf("No se puede escribir en el archivo '%s'", filePath))
		os.Exit(1)
	}

	logs.Chapter(fmt.Sprintf("El archivo '%s' ha sido creado exitosamente.\n", filePath))
}

func GetTemplateFilePath(templateFileName string) string {
	// Obtener la ruta del directorio del código actual
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Construir la ruta completa del directorio de plantillas
	templatesDir := filepath.Join(currentDir, "../../templates")

	var templateFilePath string
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == templateFileName {
			templateFilePath = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		logs.Error(fmt.Sprintf("Error al buscar el archivo de plantilla '%s'", templateFileName))
		os.Exit(1)
	}

	if templateFilePath == "" {
		msg := fmt.Sprintf("No se encontró el archivo de plantilla '%s' en el directorio '%s'.\n", templateFileName, currentDir)
		logs.Error(msg)
		os.Exit(1)
	}

	return templateFilePath
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true // El archivo existe
	}
	if os.IsNotExist(err) {
		return false // El archivo no existe
	}
	return false // Ocurrió un error al verificar la existencia del archivo
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func Prompt(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()
	if err != nil {
		logs.Error(fmt.Sprintf("Error al recibir la entrada del usuario: %s", err))
		os.Exit(1)
	}

	return result
}

func ReadFileLines(filePath string) ([]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	return lines, nil
}

func WriteFileLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
