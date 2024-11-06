from .utils import remove_env, get_env_base_dir


def main(args):
    envs = [file.stem for file in get_env_base_dir().iterdir()]

    if len(envs) == 0:
        print("Nothing to do here")
        return 1

    user_input = input("Are you sure you want to delete all environments [y/n]\n> ")
    if user_input.lower().strip() != "y":
        return 1

    for file in get_env_base_dir().iterdir():
        remove_env(file.stem)
        print(f"Environment {file.stem} removed")

    return 0
