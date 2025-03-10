from .utils import Config, cprint, get_running_containers


def main(args):
    if not Config.paths.base_dir.exists():
        return 0

    running_containers = get_running_containers()

    for env_dir in sorted(Config.paths.base_dir.iterdir()):
        if env_dir.is_dir():
            container_name = Config.templates.container_name.format(
                env_name=env_dir.stem
            )
            color = "GREEN" if container_name in running_containers else None
            cprint(f"- {env_dir.stem}", color)

    return 0
