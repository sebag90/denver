from pathlib import Path
import subprocess

from .utils import Config, cprint


ROOT = Path(__file__).parent.parent.resolve()


def main(args):
    compose_file = Path(f"{Config.paths.base_dir}/{args.name}/docker-compose.yml")
    if not compose_file.exists():
        cprint(f"Environment {args.name} is missing, create it first", "fail")
        return 1

    running_containers = subprocess.run(
        ["docker", "ps", "--format", "{{.Names}}"], capture_output=True, text=True
    )

    container_name = f"denver_{args.name}"

    # restart container if not running
    if container_name not in running_containers.stdout:
        subprocess.run(
            [
                "docker",
                "compose",
                "-f",
                compose_file,
                "up",
                "-d",
            ]
        )

    command_args_init = ["docker", "exec", "-it"]
    to_add = [container_name, "bash"]

    if args.root is True:
        to_add = ["-u", "0:0"] + to_add

    subprocess.run(command_args_init + to_add)
    return 0
