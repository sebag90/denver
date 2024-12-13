import subprocess

from .utils import Config, docker_compose, cprint


def main(args):
    # remove old image
    docker_compose(args.name, "down")
    cprint(f"Environment {args.name} stopped", "warning")
    container_tool = Config.get_config()["containers"]["container_tool"].split()
    container_args = ["image", "rm", f"{args.name}-denver_{args.name}"]
    subprocess.run(container_tool + container_args)

    # rebuild image
    docker_compose(args.name, "build")
    cprint(f"Environment {args.name} was rebuilt", "success")
    return 0
