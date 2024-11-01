from .utils import stop_docker


def main(args):
    stop_docker(args.name)
    print(f"Environment {args.name} stopped")
