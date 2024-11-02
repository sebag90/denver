from .utils import docker_compose


def main(args):
    docker_compose(args.name, "down")
    print(f"Environment {args.name} stopped")
