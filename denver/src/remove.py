from .utils import Config, remove_env, cprint


def main(args):
    envs = [file.stem for file in Config.paths.base_dir.iterdir() if file.is_dir()]
    if len(envs) == 0:
        cprint("You have no environments", "fail")
        return 1

    if args.all is False:
        for env_name in args.name:
            if env_name not in envs:
                cprint(f"Environment {env_name} not found", "fail")
            else:
                remove_env(env_name)
                cprint(f"Environment {env_name} removed", "success")

        return 0

    else:
        for file in Config.paths.base_dir.iterdir():
            if file.is_dir():
                remove_env(file.stem)
                cprint(f"Environment {file.stem} removed", "success")

        return 0
