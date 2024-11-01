from pathlib import Path
import subprocess

from .utils import get_env_base_dir


ROOT = Path(__file__).parent.parent.resolve()


def main(args):
    denver_base_dir = get_env_base_dir()
    subprocess.run(
        [
            "docker",
            "compose",
            "-f",
            Path(f"{denver_base_dir}/{args.name}/compose.yml"),
            "up",
            "-d",
        ]
    )

    subprocess.run(["docker", "exec", "-it", f"denver_{args.name}", "bash"])
