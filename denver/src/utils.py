from configparser import ConfigParser
import os
from pathlib import Path
from shutil import which
import subprocess

from pick import pick


ROOT = Path(__file__).parent.parent.resolve()


class ContainerTool:
    def __init__(self, name):
        self.name = name

    def __str__(self):
        return self.name

    @property
    def exec(self):
        return which(self.name)

    @property
    def compose(self):
        if self.name == "docker":
            return f"{self.exec} compose"

        return which(f"{self.name}-compose")


class Config:
    class paths:
        config_file = Path(f"{Path().home()}/.denver/config.ini")
        template_dir = Path(f"{ROOT}/template")
        base_dir = Path(f"{Path().home()}/.denver")
        base_dir.mkdir(exist_ok=True, parents=True)

    class templates:
        image_name = "denver/{env_name}"
        container_name = "denver_{env_name}"

    def get_config():
        Config.create_config()

        config = ConfigParser()
        config.read(Config.paths.config_file)

        return config

    def create_config():
        if Config.paths.config_file.exists():
            return

        editor_name = os.getenv("EDITOR")

        if editor_name is None or which(editor_name) is None:
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

        # choose container tool
        available = []
        for tool in [ContainerTool("docker"), ContainerTool("podman")]:
            if tool.exec is not None:
                available.append(tool)

        index = 0
        if len(available) > 1:
            container_tool, index = pick(
                [i for i in available],
                "Which tool would you like to use to manage containers?",
                indicator=">",
            )

        if len(available) > 0:
            tool = available[index]
            container_exec = tool.exec
            compose_exec = tool.compose

        else:
            cprint(
                message="No tool found to manage containers. Install podman or docker and update the config file",
                color="WARNING",
            )
            container_exec = None
            compose_exec = None

        # create config
        config = ConfigParser()
        config["general"] = {
            "editor": str(editor_path),
        }
        config["containers"] = {
            "work_dir": "workspace",
            "container_tool": container_exec,
            "compose_tool": compose_exec,
        }

        with Config.paths.config_file.open("w", encoding="utf-8") as ofile:
            config.write(ofile)


class Colors:
    PURPLE = "\033[95m"
    CYAN = "\033[96m"
    DARKCYAN = "\033[36m"
    BLUE = "\033[94m"
    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    RED = "\033[91m"
    BOLD = "\033[1m"
    UNDERLINE = "\033[4m"
    END = "\033[0m"

    def __init__(self):
        self.SUCCESS = self.GREEN
        self.WARNING = self.YELLOW
        self.FAIL = self.RED


def cprint(message, color):
    if color is None:
        print(message)
        return

    colors = Colors()
    color = getattr(colors, color)
    print(f"{color}{message}{Colors.END}")


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
        compose_tool = Config.get_config()["containers"]["compose_tool"].split()
        args = ["-f", Path(f"{env_dir}/docker-compose.yml"), action]
        subprocess.run(compose_tool + args)


def remove_env(name):
    config = Config.get_config()
    env_dir = Path(f"{Config.paths.base_dir}/{name}")

    if env_dir.exists():
        cprint(f"Stopping container {name}...", "WARNING")
        docker_compose(name, "down")

    container_tool = config["containers"]["container_tool"].split()
    image_name = Config.templates.image_name.format(env_name=name)
    args = ["image", "rm", image_name]
    subprocess.run(container_tool + args)

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


def get_running_containers():
    config = Config.get_config()

    container_args = [
        config["containers"]["container_tool"],
        "ps",
        "--format",
        "{{.Names}}",
    ]

    return set(
        subprocess.run(container_args, capture_output=True, text=True)
        .stdout.strip()
        .split()
    )
