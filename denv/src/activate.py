from pathlib import Path


ROOT = Path(__file__).parent.parent.resolve()


def main(args):
    print(Path(f"{ROOT}/template/bashrc").read_text())
