import os
from pathlib import Path
from shutil import which
import subprocess
import tomllib

from pick import pick


def get_config():
    ROOT = Path(__file__).parent.parent.resolve()
    return tomllib.loads(Path(f"{ROOT}/config.toml").read_text())


def get_env_base_dir():
    config = get_config()
    if "venv_dir" in config:
        return Path(config["venv_dir"])
    else:
        home_dir = Path().home()
        return Path(f"{home_dir}/.denver")


def rm_tree(pth):
    for child in pth.iterdir():
        if child.is_file():
            child.unlink()
        else:
            rm_tree(child)
    pth.rmdir()


def docker_compose(name, action):
    denver_base_dir = get_env_base_dir()
    env_dir = Path(f"{denver_base_dir}/{name}")

    if env_dir.exists():
        subprocess.run(
            [
                "docker",
                "compose",
                "-f",
                Path(f"{env_dir}/docker-compose.yml"),
                action,
            ]
        )


def remove_env(name):
    denver_base_dir = get_env_base_dir()
    env_dir = Path(f"{denver_base_dir}/{name}")

    if env_dir.exists():
        docker_compose(name, "down")

    subprocess.run(["docker", "image", "rm", f"{name}-denver_{name}"])

    if env_dir.exists():
        rm_tree(env_dir)


def get_editor():
    # modify the file with the standard editor
    editor = os.getenv("EDITOR", "nano")

    if editor is None or Path(which(editor)).exists() is not True:
        raise FileNotFoundError(
            "Missing text editor, set the EDITOR variable in your shell: export EDITOR=<editor of your choice>"
        )

    return editor


def modify_file(env_base_dir):
    editor = get_editor()

    while True:
        options = sorted([i.name for i in env_base_dir.iterdir()]) + ["** exit **"]
        option, index = pick(
            options, "Which file would you like to modify?", indicator=">"
        )

        if index == len(options) - 1:
            return

        subprocess.run([editor, f"{env_base_dir}/{option}"])
