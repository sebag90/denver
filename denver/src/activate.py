from pathlib import Path
import subprocess

from .utils import get_env_base_dir


ROOT = Path(__file__).parent.parent.resolve()


def main(args):
    denver_base_dir = get_env_base_dir()
    compose_file = Path(f"{denver_base_dir}/{args.name}/compose")
    if not compose_file.exists():
        print("Environment {args.name} is missing, create it first")
        return

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

    subprocess.run(["docker", "exec", "-it", container_name, "bash"])
