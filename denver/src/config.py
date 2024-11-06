from io import TextIOWrapper
import os
from pathlib import Path
import subprocess
import sys

from .utils import get_env_base_dir


def main(args):
    denver_base_dir = get_env_base_dir()
    if not Path(f"{denver_base_dir}/{args.environment}").exists():
        print(f"Environment {args.environment} is missing, create it first")
        return 1

    file_to_modify = Path(f"{denver_base_dir}/{args.environment}/{args.file}")

    # pipe stdin to the config file to modify
    if args.from_stdin is True:
        input_stream = TextIOWrapper(sys.stdin.buffer, encoding="utf-8")
        file_to_modify.write_text(input_stream.read(), encoding="utf-8")
        return 0

    # modify the file with the standard editor
    editor = os.getenv("EDITOR", "/usr/bin/nano")
    if editor is None or Path(editor).exists() is not True:
        print(
            "Missing text editor, set the EDITOR variable in your shell: export EDITOR=<editor of your choice>"
        )
        return 1

    subprocess.run([editor, file_to_modify])
    return 0
