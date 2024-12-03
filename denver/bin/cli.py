from argparse import ArgumentParser
import sys


def get_args():
    parser = ArgumentParser(
        description="Manage virtual environments with Docker", prog="denver"
    )
    subparsers = parser.add_subparsers(dest="subparser")

    # SIMPLE SUBPARSERS
    subparsers.add_parser("list", help="list all available environments")

    # CREATE ENV
    create = subparsers.add_parser("create", help="create a new environment")
    create.add_argument("name", help="the name of the environment")
    create.add_argument("--version", "-v", default="3.12", type=str)
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
    activate.add_argument(
        "-r",
        "--root",
        action="store_true",
        help="run container as root",
    )

    # REMOVE
    remove = subparsers.add_parser("remove", help="remove an environment")
    group = remove.add_mutually_exclusive_group()
    group.add_argument("name", help="the name of the environment", nargs="?")
    group.add_argument("--all", action="store_true", help="delete all environments")

    # STOP
    stop = subparsers.add_parser("stop", help="stop the container of an environment")
    stop.add_argument("name", help="the name of the environment")

    # CONFIG
    config = subparsers.add_parser(
        "config", help="modify the configuration of an environment or denver itself"
    )
    config.add_argument("name", help="the name of the environment", nargs="?")

    args = parser.parse_args()
    if args.subparser not in subparsers.choices.keys():
        parser.print_help()

    if args.subparser == "remove":
        if args.name is None and args.all is False:
            remove.print_help()
            sys.exit()

    return args
