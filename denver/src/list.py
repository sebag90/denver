import subprocess

from .utils import Config


def main(args):
    if not Config.paths.base_dir.exists():
        return 0

    running_containers = set(
        subprocess.run(
            ["docker", "ps", "--format", "{{.Names}}"], capture_output=True, text=True
        )
        .stdout.strip()
        .split()
    )

    for env_dir in sorted(Config.paths.base_dir.iterdir()):
        if env_dir.is_dir():
            prefix = "*" if f"denver_{env_dir.stem}" in running_containers else "-"
            print(f"{prefix} {env_dir.stem}")

    return 0
