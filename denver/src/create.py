from pathlib import Path

from .utils import get_env_base_dir, remove_env


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
            return

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
