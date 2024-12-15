import subprocess

from .utils import Config, docker_compose, cprint


def main(args):
    config = Config.get_config()
    # remove old image
    docker_compose(args.name, "down")
    cprint(f"Environment {args.name} stopped", "warning")
    container_tool = config["containers"]["container_tool"].split()
    image_name = Config.templates.image_name.format(env_name=args.name)

    container_args = ["image", "rm", image_name]
    subprocess.run(container_tool + container_args)

    # rebuild image
    docker_compose(args.name, "build")
    cprint(f"Environment {args.name} was built", "success")
    return 0
