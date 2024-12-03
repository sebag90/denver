import os
from pathlib import Path

from .utils import get_env_base_dir, remove_env, modify_file


ROOT = Path(__file__).parent.parent.resolve()


def main(args):
    denver_dir = get_env_base_dir()
    denver_dir.mkdir(exist_ok=True)
    new_env_dir = Path(f"{denver_dir}/{args.name}")

    if new_env_dir.exists():
        user_input = input(
            f"{args.name} exists already, do you want to replace it? [y/n]\n> "
        )
        if user_input.lower().strip() != "y":
            return 1

        remove_env(args.name)

    new_env_dir.mkdir(parents=True)

    # copy files from template
    template_vars = {
        "version": args.version,
        "name": args.name,
        "username": "root" if args.root is True else "devuser",
        "user_uid": str(os.getuid()),
        "user_gid": str(os.getgid()),
    }

    for file in Path(f"{ROOT}/template").iterdir():
        templated_file = file.read_text(encoding="utf-8")
        for var, value in template_vars.items():
            to_replace = "{{" + var + "}}"
            templated_file = templated_file.replace(to_replace, value)

        # hide .bashrc and .env
        if file.name in {"bashrc", "env"}:
            file_name = f".{file.name}"
        else:
            file_name = file.name

        Path(f"{new_env_dir}/{file_name}").write_text(templated_file, encoding="utf-8")

    # create dockerfile
    no_root_dockerfile = Path(f"{new_env_dir}/dockerfile.no_root")
    no_root_instructions = no_root_dockerfile.read_text(encoding="utf-8")
    no_root_dockerfile.unlink()

    if args.root is True:
        no_root_instructions = ""

    # update dockerfile based on user
    dockerfile_txt = Path(f"{new_env_dir}/dockerfile").read_text(encoding="utf-8")
    dockerfile_txt = dockerfile_txt.replace("{{NO_ROOT}}", no_root_instructions)
    Path(f"{new_env_dir}/dockerfile").write_text(dockerfile_txt, encoding="utf-8")

    if args.interactive is True:
        modify_file(new_env_dir)

    print(f"Environment {args.name} was created")
    return 0
