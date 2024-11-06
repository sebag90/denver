import subprocess

from .utils import docker_compose


def main(args):
    # remove old image
    docker_compose(args.name, "down")
    print(f"Environment {args.name} stopped")
    subprocess.run(["docker", "image", "rm", f"{args.name}-denver_{args.name}"])

    # rebuild image
    docker_compose(args.name, "build")
    return 0
