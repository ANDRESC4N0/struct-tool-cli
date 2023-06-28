#!/usr/bin/env python3

import os
import utils.logs as logs
import sys
import click

api_dir = os.path.join(os.getcwd(), "cmd", "api")

def to_camel_case(text):
    if not text:
        return text
    # Reemplazar espacios y caracteres especiales por guiones bajos
    text = text.strip().replace(' ', '_')
    text = ''.join(c if c.isalnum() else '_' for c in text)

    # Convertir a camel case
    words = text.split('_')
    result = words[0].lower() + ''.join(word.title() if word != words[0] else word for word in words[1:])
    return result

def to_pascal_case(text):
    if not text:
        return text
    # Reemplazar espacios y caracteres especiales por guiones bajos
    text = text.strip().replace(' ', '_')
    text = ''.join(c if c.isalnum() else '_' for c in text)

    # Convertir a Pascal case
    words = text.split('_')
    result = ''.join(word.title() for word in words)
    return result[0].upper() + result[1:]

def to_snake_case(text):
    if not text:
        return text
    # Reemplazar espacios y caracteres especiales por guiones bajos
    text = text.strip().replace(' ', '_')
    text = ''.join(c if c.isalnum() else '_' for c in text)

    # Convertir a minúsculas y agregar guiones bajos antes de letras mayúsculas
    result = ''
    for i, c in enumerate(text):
        if i > 0 and c.isupper() and text[i-1] != '_':
            result += '_'
        result += c.lower()

    return result

def create_file(file_path, template_file_path, component_name):
    name_camel_case = to_pascal_case(component_name)

    with open(template_file_path, "r") as template_file:
        content = template_file.read()
        if "NamePackage" in content:
            content = content.replace("{NamePackage}", component_name)
        if "NameComponent" in content:
            content = content.replace("{NameComponent}", name_camel_case)
        if "NameGateway" in content:
            content = content.replace("{NameGateway}", name_camel_case)
            name_function = click.prompt("Ingrese el nombre de la función", type=str)
            if name_function is not None:
                content = content.replace("{NameFunction}", name_function)

            model_response = click.prompt("Ingrese el nombre del modelo de respuesta", type=str)
            if model_response is not None:
                content = content.replace("{NameModelResponse}", model_response)

            model_request = click.prompt("Ingrese el nombre del modelo de petición", type=str)
            if model_request is not None:
                content = content.replace("{NameModelRequest}", model_request)
            


        with open(file_path, "w") as file:
            file.write(content)

    logs.chapter(f"El archivo {file_path} ha sido creado exitosamente.")

def get_template_file_path(template_file_name):
    start_dir = os.path.abspath(os.path.dirname(__file__))
    start_dir = os.path.dirname(start_dir)
    template_file_path = None

    for root, dirs, files in os.walk(start_dir):
        if template_file_name in files:
            template_file_path = os.path.join(root, template_file_name)
            break

    if template_file_path is not None:
        # El archivo se encontró en la ruta especificada en 'template_file_path'
        return template_file_path
    else:
        logs.echo(f"No se encontró el archivo {template_file_name} en el directorio {start_dir}.")
        sys.exit(1)