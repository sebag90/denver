from pathlib import Path

from .utils import get_env_base_dir, modify_menu, modify_single_file, CONFIG_FILE


def main(args):
    if args.name is None:
        modify_single_file(CONFIG_FILE)
        return 0

    denver_base_dir = get_env_base_dir()
    if not Path(f"{denver_base_dir}/{args.name}").exists():
        print(f"Environment {args.name} is missing, create it first")
        return 1

    modify_menu(Path(f"{denver_base_dir}/{args.name}"))

    return 0
