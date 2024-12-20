import subprocess

from .utils import Config, cprint, get_running_containers


def main(args):
    running_containers = get_running_containers()
    container_name = Config.templates.container_name.format(env_name=args.name)
    image_name = Config.templates.image_name.format(env_name=args.name)

    if container_name not in running_containers:
        cprint(f"Image for {args.name} is missing, build it first", "FAIL")
        return 1

    config = Config.get_config()
    subprocess.run(
        [config["containers"]["container_tool"], "commit", container_name, image_name]
    )

    cprint(
        f"The image {image_name} was updated based on container {container_name}",
        "SUCCESS",
    )
    return 0
