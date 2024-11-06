from .utils import remove_env


def main(args):
    user_input = input(f"Are you sure you want to delete {args.name} [y/n]\n> ")
    if user_input.lower().strip() != "y":
        return 1

    remove_env(args.name)
    print(f"Environment {args.name} removed")
    return 0
