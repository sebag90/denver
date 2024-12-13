from pathlib import Path
import subprocess

from .utils import Config, cprint


ROOT = Path(__file__).parent.parent.resolve()


def main(args):
    compose_file = Path(f"{Config.paths.base_dir}/{args.name}/docker-compose.yml")
    if not compose_file.exists():
        cprint(f"Environment {args.name} is missing, create it first", "fail")
        return 1

    container_tool = Config.get_config()["containers"]["container_tool"].split()
    container_args = ["ps", "--format", "{{.Names}}"]
    running_containers = subprocess.run(
        container_tool + container_args, capture_output=True, text=True
    )

    container_name = f"denver_{args.name}"

    # restart container if not running
    if container_name not in running_containers.stdout:
        compose_tool = Config.get_config()["containers"]["compose_tool"].split()
        container_args = ["-f", compose_file, "up", "-d"]
        subprocess.run(compose_tool + container_args)

    command_args_init = container_tool + ["exec", "-it"]
    container_args = [container_name, "bash"]

    if args.root is True:
        container_args = ["-u", "0:0"] + container_args

    subprocess.run(command_args_init + container_args)
    return 0
