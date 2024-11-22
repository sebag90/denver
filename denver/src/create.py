from pathlib import Path
import subprocess

from pick import pick

from .utils import get_env_base_dir, remove_env, get_editor


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
    template_vars = {"version": args.version, "name": args.name}

    for file in Path(f"{ROOT}/template").iterdir():
        templated_file = file.read_text(encoding="utf-8")
        for var, value in template_vars.items():
            to_replace = "{{" + var + "}}"
            templated_file = templated_file.replace(to_replace, value)

        Path(f"{new_env_dir}/{file.name}").write_text(templated_file, encoding="utf-8")

    # pick dockerfile based on root
    dockerfile = Path(f"{new_env_dir}/dockerfile.root")
    if args.root is True:
        dockerfile.replace(dockerfile.with_suffix(""))
    else:
        dockerfile.unlink()

    if args.interactive is True:
        editor = get_editor()

        while True:
            options = [i.stem for i in new_env_dir.iterdir() if i.stem != "bashrc"] + [
                "exit"
            ]
            option, index = pick(
                options, "Which file would you like to modify?", indicator=">"
            )

            if index == len(options) - 1:
                break

            subprocess.run([editor, f"{new_env_dir}/{option}"])

    print(f"Environment {args.name} was created")
    return 0
