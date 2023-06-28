#!/usr/bin/env python3

import os
import sys
import re
import click
import utils.utils as utils
import utils.logs as logs
import injection_wire_app as injection

api_dir = os.path.join(os.getcwd(), "cmd", "api")

@click.command()
@click.option("--component-name", "-c", prompt='Ingrese el nombre del componente', help="Nombre del componente")
@click.option("--service-name", "-s", prompt='Ingrese el nombre del servicio', help="Nombre del servicio")
def add_service_component(component_name, service_name):
    """Agrega un servicio a un componente"""
    add_service(component_name, service_name)

def add_service(component_name, service_name):
    if not component_name:
        component_name = click.prompt("Ingrese el nombre del componente").strip()
    if not service_name:
        service_name = click.prompt("Ingrese el nombre del servicio").strip()

    # Construir la ruta del archivo del servicio
    api_components_dir = os.path.join(api_dir, "components")
    service_file_path = os.path.join(api_components_dir, component_name, "service.go")

    # Verificar si el archivo del servicio existe
    if not os.path.exists(service_file_path):
        logs.error(f"Error: El archivo '{service_file_path}' no existe en la ruta actual.")
        sys.exit(1)
    
    # Abrir el archivo en modo de lectura y escritura
    with open(service_file_path, "r+") as file:
        lines = file.readlines()
        pascal = utils.to_pascal_case(component_name)

        # Buscar la línea donde se debe agregar el nuevo contenido
        target_line = "type I{NameComponent}Service interface {".replace("{NameComponent}", pascal)
        for i, line in enumerate(lines):
            if target_line in line:
                # Insertar el nuevo contenido después de la línea objetivo
                lines.insert(i + 1, f"\t{service_name}(requestSearch interface{{}}, ctx context.Context, txn *newrelic.Transaction) (interface{{}}, error)\n")
                break
        else:
            logs.error(f"Error: No se encontró la línea '{target_line}' en el archivo '{service_file_path}'.")
            sys.exit(1)


        # Volver al principio del archivo y escribir las líneas modificadas
        lines.append("\n" + f"func (service *{pascal}Service) {service_name}(requestSearch interface{{}}, ctx context.Context, txn *newrelic.Transaction) (interface{{}}, error){{\n    return nil, nil\n}}" + "\n")
        file.seek(0)
        file.writelines(lines)
        file.truncate()
        file.close()

    logs.chapter(f"El contenido ha sido agregado exitosamente en el archivo '{service_file_path}'.")

def injection_services(api_components_dir, component_name, service_interfaces, gateway_interfaces) -> object:
    result = {}
    list_services = []

    if service_interfaces:
        logs.echo("Se encontraron los siguientes nombres de interfaces de servicio:")
        click.echo(list(service_interfaces.keys()))
        list_comp_service = click.prompt("Ingrese los servicios que requieres en el componente", default="").strip()
        if list_comp_service:
            list_services.extend(list_comp_service.split(" "))
    else:
        logs.echo("No se encontraron nombres de interfaces de servicio.")

    if gateway_interfaces:
        logs.echo("Se encontraron los siguientes nombres de interfaces de gateway:")
        click.echo(list(gateway_interfaces.keys()))
        list_gateway = click.prompt("Ingrese los servicios que requieres en el componente", default="").strip()
        if list_gateway:
            list_services.extend(list_gateway.split(" "))
    else:
        logs.echo("No se encontraron nombres de interfaces de gateway.")

    # Agregar los servicios al archivo service.go
    service_file_path = os.path.join(api_components_dir, component_name, "service.go")
    with open(service_file_path, "r+") as file:
        lines = file.readlines()

        # Buscar la línea donde se debe agregar el nuevo contenido
        target_line = f"type {utils.to_pascal_case(component_name)}Service struct {{\n"
        for i, line in enumerate(lines):
            if line.strip() == target_line.strip():
                # Insertar el nuevo contenido después de la línea objetivo
                for service_name in list_services:
                    package_service = ""
                    if service_name in service_interfaces:
                        package_service = service_interfaces.get(service_name, {}).get("package") + f".{service_name}Service"
                    elif service_name in gateway_interfaces:
                        package_service = gateway_interfaces.get(service_name, {}).get("package") + f".{service_name}Gateway"
                    
                    if package_service:
                        result[service_name[1:]] = package_service
                        lines.insert(i + 1, f"\t{service_name[1:]} {package_service}\n")
                break
        else:
            logs.error(f"Error: No se encontró la línea '{target_line.strip()}' en el archivo '{service_file_path}'.")
            sys.exit(1)

        file.seek(0)
        file.writelines(lines)
        file.truncate()

    logs.chapter(f"Se agregaron los servicios al archivo '{service_file_path}'.")

    return result

def find_interfaces(directory, pattern):
    if not os.path.exists(directory) or not os.path.isdir(directory):
        logs.echo(f"El directorio {directory} no existe.")
        return {}

    interfaces = {}

    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith(".go"):
                file_path = os.path.join(root, file)
                with open(file_path, "r") as f:
                    content = f.read()
                    matches = re.findall(pattern, content)
                    if matches:
                        for match in matches:
                            interface_name = f"I{match}"
                            package = os.path.basename(os.path.dirname(file_path))
                            interfaces[interface_name] = {"package": package}

    return interfaces

@click.command()
@click.option('--component_name', prompt='Ingrese el nombre del componente', help='El nombre del componente a crear.')
def create_component(component_name):
    """Crea un componente en la ruta 'cmd/api'."""
    
    # Construir la ruta del directorio del componente dentro de 'cmd/api'
    api_components_dir = os.path.join(api_dir, "components")
    component_dir = os.path.join(api_components_dir, component_name)

    # Verificar si la ruta 'cmd/api' existe
    if not os.path.exists(api_components_dir):
        logs.error("Error: La ruta 'cmd/api' no existe.")
        raise

    # Verificar si la ruta del componente ya existe
    if os.path.exists(component_dir):
        logs.error(f"Error: El componente '{component_name}' ya existe en la ruta 'cmd/api'.")
        raise

    # Crear el directorio del componente dentro de 'cmd/api'
    os.makedirs(component_dir)

    service_file_path = os.path.join(component_dir, "service.go")
    controller_file_path = os.path.join(component_dir, "controller.go")

    service_template_file_path = utils.get_template_file_path("service_template.txt")
    controller_template_file_path = utils.get_template_file_path("controller_template.txt")

    utils.create_file(service_file_path, service_template_file_path, component_name)
    utils.create_file(controller_file_path, controller_template_file_path, component_name)

    logs.chapter(f"El componente '{component_name}' ha sido creado exitosamente en la ruta 'cmd/api'.")

    # Obtener las interfaces de servicios y gateways
    service_interfaces = find_interfaces(api_components_dir, r"I(\w+)Service interface")
    gateway_interfaces = find_interfaces(os.path.join(api_dir, "gateway"), r"I(\w+)Gateway interface")

    # Inyectar los servicios en el archivo service.go
    injections = injection_services(api_components_dir, component_name, service_interfaces, gateway_interfaces)
    provider_file_path = os.path.join(api_dir, "app/providers/components.go")
    provider_template_file_path = utils.get_template_file_path("provider_components_template.txt")
    
    if not os.path.exists(provider_file_path):
        utils.create_file(provider_file_path, provider_template_file_path, component_name)
    
    with open(provider_file_path, 'a') as file:
        name_pascal = utils.to_pascal_case(component_name)
        file.write(f"\n\nfunc Init{name_pascal}Service(\n")
        logs.echo(f"Inyecciones: {injections}")
        for key, value in injections.items():
            camel = utils.to_camel_case(key)
            file.write(f"\t{camel} *{value},\n")
        file.write(f") *{component_name}.{name_pascal}Service {{\n\treturn &{component_name}.{name_pascal}Service{{\n")
        for key, _ in injections.items():
            camel = utils.to_camel_case(key)
            file.write(f"\t\t{key}: {camel},\n")
        file.write("\t}\n}\n\n")
    
    # TODO: Inyectar el componente en el archivo 'app/app.go' y ejecutar 'go generate ./...'
    injection.add_injection_app(component_name, "components")

    # add_service(component_name, 'ExampleService')
