from pathlib import Path
import subprocess

from .utils import get_env_base_dir


def rm_tree(pth):
    for child in pth.iterdir():
        if child.is_file():
            child.unlink()
        else:
            rm_tree(child)
    pth.rmdir()


def remove_env(name):
    denver_base_dir = get_env_base_dir()
    env_dir = Path(f"{denver_base_dir}/{name}")

    if env_dir.exists():
        subprocess.run(
            [
                "docker",
                "compose",
                "-f",
                Path(f"{env_dir}/compose.yml"),
                "down",
            ]
        )
    try:
        subprocess.run(["docker", "image", "rm", f"{name}-denver_{name}"])
    except FileNotFoundError:
        pass

    if env_dir.exists():
        rm_tree(env_dir)


def main(args):
    user_input = input(f"Are you sure you want to delete {args.name} [y/n]\n> ")

    if user_input.lower().strip() != "y":
        return

    remove_env(args.name)

    print(f"Environment {args.name} removed")
