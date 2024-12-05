from pathlib import Path

from .utils import Config, modify_menu, modify_single_file


def main(args):
    if args.name is None:
        modify_single_file(Config.paths.config_file)
        return 0

    if not Path(f"{Config.paths.base_dir}/{args.name}").exists():
        print(f"Environment {args.name} is missing, create it first")
        return 1

    modify_menu(Path(f"{Config.paths.base_dir}/{args.name}"))

    return 0
