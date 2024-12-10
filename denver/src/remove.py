from .utils import Config, remove_env


def main(args):
    envs = [file.stem for file in Config.paths.base_dir.iterdir() if file.is_dir()]
    if len(envs) == 0:
        print("You have no environments")
        return 1

    if args.all is False:
        for env_name in args.name:
            if env_name not in envs:
                print(f"Environment {env_name} not found")
            else:
                remove_env(env_name)
                print(f"Environment {env_name} removed")

        return 0

    else:
        for file in Config.paths.base_dir.iterdir():
            if file.is_dir():
                remove_env(file.stem)
                print(f"Environment {file.stem} removed")

        return 0
