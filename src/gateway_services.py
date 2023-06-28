#!/usr/bin/env python3

import os
import sys
import click
import utils.utils as utils
import utils.logs as logs

api_dir = os.path.join(os.getcwd(), "cmd", "api")

@click.command()
@click.option("--name", "-n", prompt="Ingresa el nombre del gateway", help="Nombre del gateway", required=True)
def create_gateway(name, rest):
    """Crea un nuevo gateway"""
    click.echo(f'Se seleccionó el método REST: {rest}')

    gateway_dir = os.path.join(api_dir, "gateway")
    # Verificar si la ruta 'cmd/api/gateway' existe
    if not os.path.exists(gateway_dir):
        logs.info("Creando la ruta 'cmd/api/gateway'...")
        os.makedirs(gateway_dir)
    
    # Verificar si el gateway ya existe
    gateway_file = os.path.join(gateway_dir, name + ".go")
    if os.path.exists(gateway_file):
        logs.error(f"El gateway '{name}' ya existe.")
        sys.exit(1)

    logs.echo(f"Creando el gateway {gateway_file}...")
    gateway_template_file_path = utils.get_template_file_path("gateway_template.txt")
    utils.create_file(gateway_file, gateway_template_file_path, name)

    logs.echo(f"Gateway {name} creado exitosamente.")

@click.command()
@click.option('--rest', type=click.Choice(['GET', 'POST', 'PUT', 'DELETE']), prompt='Selecciona una opción de método REST', help='Opciones: GET, POST, PUT, DELETE')
def add_service_gateway(rest):
    """Agrega un servicio al gateway"""
    click.echo(f'Se seleccionó el método REST: {rest}')
