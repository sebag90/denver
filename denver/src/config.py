from pathlib import Path

from .utils import get_env_base_dir, modify_file


def main(args):
    denver_base_dir = get_env_base_dir()
    if not Path(f"{denver_base_dir}/{args.name}").exists():
        print(f"Environment {args.name} is missing, create it first")
        return 1

    modify_file(Path(f"{denver_base_dir}/{args.name}"))

    return 0
