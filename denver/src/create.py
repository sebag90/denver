import os
from pathlib import Path
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
        # mac behaves differently than linux, use standard 1000 for mac
        "user_uid": str(os.getuid()) if sys.platform != "darwin" else "1000",
        "user_gid": str(os.getgid()) if sys.platform != "darwin" else "1000",
    }

    for file in Config.paths.template_dir.iterdir():
        templated_file = file.read_text(encoding="utf-8")
        for var, value in template_vars.items():
            to_replace = "{{" + var + "}}"
            templated_file = templated_file.replace(to_replace, value)

        # hide .bashrc and .env
        if file.name in {"bashrc", "env"}:
            file_name = f".{file.name}"
        else:
            file_name = file.name

        Path(f"{new_env_dir}/{file_name}").write_text(templated_file, encoding="utf-8")

    if args.interactive is True:
        modify_menu(new_env_dir)

    print(f"Environment {args.name} was created")
    return 0
