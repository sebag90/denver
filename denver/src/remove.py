from .utils import Config, remove_env


def main(args):
    envs = [file.stem for file in Config.paths.base_dir.iterdir() if file.is_dir()]
    if len(envs) == 0:
        print("You have no environments")
        return 1

    if args.name is not None:
        if args.name not in envs:
            print(f"Environment {args.name} not found")
            return 1

    env_name = "all environments" if args.all is True else args.name
    user_input = input(f"Are you sure you want to delete {env_name} [y/n]\n> ")
    if user_input.lower().strip() != "y":
        return 1

    if args.all is False:
        remove_env(args.name)
        print(f"Environment {args.name} removed")
        return 0

    else:
        for file in Config.paths.base_dir.iterdir():
            if file.is_dir():
                remove_env(file.stem)
                print(f"Environment {file.stem} removed")

        return 0
