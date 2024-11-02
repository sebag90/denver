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
        return

    file_to_modify = Path(f"{denver_base_dir}/{args.environment}/{args.file}")

    if args.from_stdin is True:
        input_stream = TextIOWrapper(sys.stdin.buffer, encoding="utf-8")
        file_to_modify.write_text(input_stream.read(), encoding="utf-8")
        return

    editor = os.getenv("EDITOR")
    subprocess.run([editor, file_to_modify])
