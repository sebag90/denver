import os
from pathlib import Path
import subprocess

from .utils import get_env_base_dir


def main(args):
    editor = os.getenv("EDITOR")
    denver_base_dir = get_env_base_dir()
    file = Path(f"{denver_base_dir}/{args.environment}/{args.file}")
    subprocess.run([editor, file])
