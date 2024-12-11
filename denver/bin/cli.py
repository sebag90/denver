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
    remove.add_argument(
        "name",
        help="the name of the environment (can not be used in combination with --all)",
        nargs="*",
        default=None,
    )
    remove.add_argument(
        "--all",
        action="store_true",
        help="delete all environments (can not be used in combination with named environment)",
    )

    # STOP
    stop = subparsers.add_parser("stop", help="stop the container of an environment")
    stop.add_argument("name", help="the name of the environment", nargs="+")

    # CONFIG
    config = subparsers.add_parser(
        "config", help="modify the configuration of an environment or denver itself"
    )
    config.add_argument("name", help="the name of the environment", nargs="?")

    args = parser.parse_args()

    def allowed_combination():
        if args.subparser not in subparsers.choices.keys():
            parser.print_help()
            return False

        if args.subparser == "remove":
            # no named env, no --all
            if not args.name and args.all is False:
                remove.print_help()
                return False

            # both named env and --all
            if args.name and args.all is True:
                remove.print_help()
                return False

        return True

    if allowed_combination() is False:
        sys.exit()

    return args
