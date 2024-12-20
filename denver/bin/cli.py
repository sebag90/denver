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

    # BUILD
    build = subparsers.add_parser("build", help="build an environment from image")
    build.add_argument("name", help="the name of the environment")

    # ACTIVATE
    shell = subparsers.add_parser("shell", help="activate the shell of an environment")
    shell.add_argument("name", help="the name of the environment")
    shell.add_argument(
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

    # RUN
    run = subparsers.add_parser(
        "run", help="run a command using the selected environment"
    )
    run.add_argument(
        "--environment", "-e", help="the enviroonment to use", required=True
    )
    run.add_argument(
        "--detach",
        "-d",
        help="detach the command from the current shell",
    )
    run.add_argument(
        "command", help="the command you want to execute", nargs="+", type=str
    )

    # UPDATE
    update = subparsers.add_parser(
        "update",
        help="update the image of an environment based on the curent state of a running container",
    )
    update.add_argument(
        "name",
        help="the name of the environment",
    )

    args, unknown = parser.parse_known_args()

    # workaround argparse limitation
    if args.subparser == "run":
        args.command += unknown

    else:
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
