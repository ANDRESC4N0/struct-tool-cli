#!/usr/bin/env python3

import os
import sys
import utils.utils as utils
import utils.logs as logs

api_dir = os.path.join(os.getcwd(), "cmd", "api")

def add_injection_app(component_name, injection_type):
    template_file_path = utils.get_template_file_path("app_template.txt")
    app_path = os.path.join(api_dir, "app/app.go")

    # Verificar si la ruta del app ya existe
    if not os.path.exists(app_path):
        logs.echo("No existe app/app.go. Creando...")
        utils.create_file(app_path, template_file_path, None)

    # Buscar para insertar nueva inyección
    with open(app_path, "r+") as file:
        lines = file.readlines()
        name_pascal_case = utils.to_pascal_case(component_name)
        target_line = f"var {injection_type}RouterSet = wire.NewSet("
        found = False
        line_position = 0

        for i, line in enumerate(lines):
            if line.strip() == target_line.strip():
                found = True
                injection_line = get_injection_line(component_name, name_pascal_case, injection_type)

                lines.insert(i + 1, injection_line)
                break

            if not found and line.strip() == ")".strip():
                line_position = i
            if not found and line.strip() == "panic(wire.Build(".strip():
                lines.insert(i + 1, f"\t{injection_type}RouterSet,\n")

        if not found and line_position > 0:
            injection_lines = get_injection_lines(component_name, name_pascal_case, injection_type)
            lines.insert(line_position + 1, f"\nvar {injection_type}RouterSet = wire.NewSet(\n")
            lines[line_position + 2:line_position + 2] = injection_lines
            lines.insert(line_position + 2 + len(injection_lines), f")\n")

        file.seek(0)
        file.writelines(lines)


def get_injection_line(component_name, name_pascal_case, injection_type):
    if injection_type == "components":
        return f"\tproviders.Init{name_pascal_case}Service, wire.Bind(new({component_name}.I{name_pascal_case}Service), new(*{component_name}.{name_pascal_case}Service)),\n"
    elif injection_type == "controllers":
        return f"\tproviders.Init{name_pascal_case}Controller, wire.Bind(new({component_name}.I{name_pascal_case}Controller), new(*{component_name}.{name_pascal_case}Controller)),\n"
    elif injection_type == "repositories":
        return f"\tproviders.Init{name_pascal_case}Repository, wire.Bind(new({component_name}.I{name_pascal_case}Repository), new(*{component_name}.{name_pascal_case}Repository)),\n"
    elif injection_type == "clients":
        return f"\tproviders.Init{name_pascal_case}Client, wire.Bind(new(clients.I{name_pascal_case}Client), new(*clients.{name_pascal_case}Client)),\n"
    elif injection_type == "gateway":
        return f"\tproviders.Init{name_pascal_case}Gateway, wire.Bind(new(gateway.I{name_pascal_case}Gateway), new(*gateway.{name_pascal_case}Gateway)),\n"
    elif injection_type == "middlewares":
        return f"\tproviders.Init{name_pascal_case}Middleware, wire.Bind(new(middlewares.I{name_pascal_case}Middleware), new(*middlewares.{name_pascal_case}Middleware)),\n"
    else:
        raise ValueError("Tipo de inyección no válido")


def get_injection_lines(component_name, name_pascal_case, injection_type):
    injection_line = get_injection_line(component_name, name_pascal_case, injection_type)
    return [injection_line]