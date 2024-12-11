import subprocess

from .utils import Config, cprint


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
            color = "green" if f"denver_{env_dir.stem}" in running_containers else None
            cprint(f"* {env_dir.stem}", color)

    return 0
