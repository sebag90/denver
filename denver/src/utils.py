from configparser import ConfigParser
import os
from pathlib import Path
from shutil import which
import subprocess

from pick import pick


ROOT = Path(__file__).parent.parent.resolve()


class Config:
    class paths:
        config_file = Path(f"{Path().home()}/.denver/config.ini")
        template_dir = Path(f"{ROOT}/template")
        base_dir = Path(f"{Path().home()}/.denver")
        base_dir.mkdir(exist_ok=True, parents=True)

    def get_config():
        Config.create_config()

        config = ConfigParser()
        config.read(Config.paths.config_file)
        return config

    def create_config():
        if Config.paths.config_file.exists():
            return

        editor_name = os.getenv("EDITOR")
        if editor_name is None or which(editor_name) is not None:
            while True:
                user_input = input(
                    "No editor found in your environment, enter your preferred editor\n> "
                )
                if which(user_input.strip()) is not None:
                    editor_name = user_input.strip()
                    print(
                        f"Your editor: {which(user_input.strip())}. You can change this with: denver config"
                    )
                    break

        editor_path = which(editor_name)

        # create config
        config = ConfigParser()
        config["general"] = {
            "editor": str(editor_path),
        }
        config["containers"] = {"work_dir": "workspace"}

        with Config.paths.config_file.open("w", encoding="utf-8") as ofile:
            config.write(ofile)


class Colors:
    header = "\033[95m"
    blue = "\033[94m"
    success = "\033[96m"
    green = "\033[92m"
    warning = "\033[93m"
    fail = "\033[91m"
    eos = "\033[0m"
    bold = "\033[1m"
    underline = "\033[4m"


def cprint(message, color):
    if color is None:
        print(message)
        return

    color = getattr(Colors, color)
    print(f"{color}{message}{Colors.eos}")


def rm_tree(pth):
    for child in pth.iterdir():
        if child.is_file():
            child.unlink()
        else:
            rm_tree(child)
    pth.rmdir()


def docker_compose(name, action):
    env_dir = Path(f"{Config.paths.base_dir}/{name}")

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
    env_dir = Path(f"{Config.paths.base_dir}/{name}")

    if env_dir.exists():
        docker_compose(name, "down")

    subprocess.run(["docker", "image", "rm", f"{name}-denver_{name}"])

    if env_dir.exists():
        rm_tree(env_dir)


def modify_single_file(file):
    editor = Config.get_config()["general"]["editor"]
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
