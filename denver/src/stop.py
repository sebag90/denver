from pathlib import Path

from .utils import Config, docker_compose


def main(args):
    for env_name in args.name:
        if not Path(f"{Config.paths.base_dir}/{env_name}").exists():
            print(f"Environment {env_name} is missing, create it first")

        else:
            docker_compose(env_name, "down")
            print(f"Environment {env_name} stopped")

    return 0
