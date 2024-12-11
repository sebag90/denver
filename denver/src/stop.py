from pathlib import Path

from .utils import Config, docker_compose, cprint


def main(args):
    for env_name in args.name:
        if not Path(f"{Config.paths.base_dir}/{env_name}").exists():
            cprint(f"Environment {env_name} is missing, create it first", "fail")

        else:
            docker_compose(env_name, "down")
            cprint(f"Environment {env_name} stopped", "success")

    return 0
