from configparser import ConfigParser
import os
from pathlib import Path
from shutil import which
import subprocess

from pick import pick


ROOT = Path(__file__).parent.parent.resolve()
CONFIG_FILE = Path(f"{ROOT}/config.ini")


def create_config():
    if CONFIG_FILE.exists():
        return

    config = ConfigParser()
    config["denver"] = {"base_dir": f"{Path().home()}/.denver"}
    config["containers"] = {"work_dir": "workspace"}

    with CONFIG_FILE.open("w", encoding="utf-8") as ofile:
        config.write(ofile)


def get_config():
    create_config()

    config = ConfigParser()
    config.read(CONFIG_FILE)
    return config


def get_env_base_dir():
    config = get_config()
    return Path(config["denver"]["base_dir"])


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


def modify_single_file(file):
    editor = get_editor()
    subprocess.run([editor, file])


def modify_menu(env_base_dir):
    while True:
        options = sorted([i.name for i in env_base_dir.iterdir()]) + ["** exit **"]
        option, index = pick(
            options, "Which file would you like to modify?", indicator=">"
        )

        if index == len(options) - 1:
            return

        modify_single_file(f"{env_base_dir}/{option}")
