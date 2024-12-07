import os
from pathlib import Path
from string import Template
import sys

from .utils import Config, remove_env, modify_menu


def main(args):
    Config.paths.base_dir.mkdir(exist_ok=True)
    new_env_dir = Path(f"{Config.paths.base_dir}/{args.name}")

    if new_env_dir.exists():
        user_input = input(
            f"{args.name} exists already, do you want to replace it? [y/n]\n> "
        )
        if user_input.lower().strip() != "y":
            return 1

        remove_env(args.name)

    new_env_dir.mkdir(parents=True)

    # copy files from template
    template_vars = {
        "version": args.version,
        "name": args.name,
        "username": "devuser",
        "work_dir": Config.get_config()["containers"]["work_dir"],
        # mac behaves differently than linux, use standard 1000 for mac
        "user_uid": str(os.getuid()) if sys.platform != "darwin" else "1000",
        "user_gid": str(os.getgid()) if sys.platform != "darwin" else "1000",
    }

    to_template = {"docker-compose.yml"}

    for file in Config.paths.template_dir.iterdir():
        file_content = file.read_text(encoding="utf-8")
        if file.name in to_template:
            file_content = Template(file_content).substitute(template_vars)

        # hide .bashrc and .env
        if file.name in {"bashrc", "env"}:
            file_name = f".{file.name}"
        else:
            file_name = file.name

        Path(f"{new_env_dir}/{file_name}").write_text(file_content, encoding="utf-8")

    if args.interactive is True:
        modify_menu(new_env_dir)

    print(f"Environment {args.name} was created")
    return 0
