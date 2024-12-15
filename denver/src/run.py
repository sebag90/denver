from pathlib import Path
import subprocess

from .utils import Config, cprint


def main(args):
    config = Config.get_config()

    container_tool = config["containers"]["container_tool"].split()
    container_args = ["image", "list", "--format", "{{.Repository}}"]
    active_images = set(
        subprocess.run(container_tool + container_args, capture_output=True, text=True)
        .stdout.strip()
        .split()
    )

    active_images = [i.replace("localhost/", "") for i in active_images]
    image_name = Config.templates.image_name.format(env_name=args.environment)

    if image_name not in active_images:
        cprint(f"Image for {args.environment} is missing, build it first", "fail")
        return 1

    executable, *run_args = args.command

    container_args_1 = [
        "run",
        "-it",
        "-v",
        f"{Path().resolve()}:/workspace",
    ]

    container_args_2 = ["--entrypoint", executable, image_name] + run_args

    if args.detach is True:
        container_args_1.append("-d")

    complete_args_list = container_args_1 + container_args_2

    subprocess.run(container_tool + complete_args_list)
