from .utils import remove_env, get_env_base_dir


def main(args):
    user_input = input("Are you sure you want to delete all environments [y/n]\n> ")
    if user_input.lower().strip() != "y":
        return

    for file in get_env_base_dir().iterdir():
        remove_env(file.stem)
        print(f"Environment {file.stem} removed")
