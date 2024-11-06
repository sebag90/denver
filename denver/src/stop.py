from pathlib import Path

from .utils import docker_compose, get_env_base_dir


def main(args):
    denver_base_dir = get_env_base_dir()
    if not Path(f"{denver_base_dir}/{args.name}").exists():
        print(f"Environment {args.name} is missing, create it first")
        return 1

    docker_compose(args.name, "down")
    print(f"Environment {args.name} stopped")

    return 0
