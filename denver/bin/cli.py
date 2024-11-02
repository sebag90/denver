from argparse import ArgumentParser


def get_args():
    parser = ArgumentParser(
        description="Manage virtual environments with Docker", prog="denv"
    )
    subparsers = parser.add_subparsers(dest="subparser")

    create = subparsers.add_parser("create", help="create a new environment")
    create.add_argument("name")
    create.add_argument("--version", "-v", default="3.12", type=str)
    create.add_argument(
        "--root", action="store_true", help="the user inside the container will be root"
    )

    subparsers.add_parser("list", help="list all available environments")
    subparsers.add_parser("prune", help="remove all environments")

    activate = subparsers.add_parser(
        "activate", help="activate the shell of an environment"
    )
    activate.add_argument("name")

    remove = subparsers.add_parser("remove", help="remove an environment")
    remove.add_argument("name")

    stop = subparsers.add_parser("stop", help="stop the container of an environment")
    stop.add_argument("name")

    config = subparsers.add_parser(
        "config", help="modify the configuration of an environment"
    )
    config.add_argument(
        "file",
        help="the file to modify",
        choices=["env", "dockerfile", "requirements", "compose"],
    )
    config.add_argument(
        "-e", "--environment", help="the environment to modify", required=True
    )

    args = parser.parse_args()
    if args.subparser not in subparsers.choices.keys():
        parser.print_help()

    return args
