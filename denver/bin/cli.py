from argparse import ArgumentParser


def get_args():
    parser = ArgumentParser(
        description="Manage virtual environments with Docker", prog="denv"
    )
    subparsers = parser.add_subparsers(dest="subparser")

    create = subparsers.add_parser("create", help="create a new environment")
    create.add_argument("name")
    create.add_argument("--version", "-v", default="3.12", type=str)

    list_envs = subparsers.add_parser("list")

    activate = subparsers.add_parser("activate")
    activate.add_argument("name")

    remove = subparsers.add_parser("remove")
    remove.add_argument("name")

    args = parser.parse_args()
    if args.subparser not in subparsers.choices.keys():
        parser.print_help()

    return args
