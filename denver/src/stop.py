from pathlib import Path

from .utils import Config, docker_compose


def main(args):
    if not Path(f"{Config.paths.base_dir}/{args.name}").exists():
        print(f"Environment {args.name} is missing, create it first")
        return 1

    docker_compose(args.name, "down")
    print(f"Environment {args.name} stopped")

    return 0
