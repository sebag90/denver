import subprocess

from .utils import Config, cprint


def main(args):
    if not Config.paths.base_dir.exists():
        return 0

    config = Config.get_config()

    container_tool = config["containers"]["container_tool"].split()
    container_args = ["ps", "--format", "{{.Names}}"]

    running_containers = set(
        subprocess.run(container_tool + container_args, capture_output=True, text=True)
        .stdout.strip()
        .split()
    )

    for env_dir in sorted(Config.paths.base_dir.iterdir()):
        if env_dir.is_dir():
            container_name = Config.templates.container_name.format(
                env_name=env_dir.stem
            )
            color = "green" if container_name in running_containers else None
            cprint(f"* {env_dir.stem}", color)

    return 0
