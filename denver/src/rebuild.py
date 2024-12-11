import subprocess

from .utils import docker_compose, cprint


def main(args):
    # remove old image
    docker_compose(args.name, "down")
    cprint(f"Environment {args.name} stopped", "warning")
    subprocess.run(["docker", "image", "rm", f"{args.name}-denver_{args.name}"])

    # rebuild image
    docker_compose(args.name, "build")
    cprint(f"Environment {args.name} was rebuilt", "success")
    return 0
