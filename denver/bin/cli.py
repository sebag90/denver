from argparse import ArgumentParser


def get_args():
    parser = ArgumentParser(
        description="Manage virtual environments with Docker", prog="denver"
    )
    subparsers = parser.add_subparsers(dest="subparser")

    # SIMPLE SUBPARSERS
    subparsers.add_parser("list", help="list all available environments")
    subparsers.add_parser("prune", help="remove all environments")

    # CREATE ENV
    create = subparsers.add_parser("create", help="create a new environment")
    create.add_argument("name", help="the name of the environment")
    create.add_argument("--version", "-v", default="3.12", type=str)
    create.add_argument(
        "-r",
        "--root",
        action="store_true",
        help="the user inside the container will be root",
    )
    create.add_argument(
        "-i",
        "--interactive",
        action="store_true",
        help="modify environment files upon creation",
    )

    # REBUILD
    rebuild = subparsers.add_parser("rebuild", help="rebuild an environment from image")
    rebuild.add_argument("name", help="the name of the environment")

    # ACTIVATE
    activate = subparsers.add_parser(
        "activate", help="activate the shell of an environment"
    )
    activate.add_argument("name", help="the name of the environment")

    # REMOVE
    remove = subparsers.add_parser("remove", help="remove an environment")
    remove.add_argument("name", help="the name of the environment")

    # STOP
    stop = subparsers.add_parser("stop", help="stop the container of an environment")
    stop.add_argument("name", help="the name of the environment")

    # CONFIG
    config = subparsers.add_parser(
        "config", help="modify the configuration of an environment"
    )
    config.add_argument("name", help="the name of the environment")

    args = parser.parse_args()
    if args.subparser not in subparsers.choices.keys():
        parser.print_help()

    return args
