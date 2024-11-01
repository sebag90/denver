import importlib
from pathlib import Path

from .cli import get_args


ROOT = Path(__file__).parent.parent.resolve()


def main():
    args = get_args()

    if args.subparser is not None:
        module = importlib.import_module(f"{ROOT.stem}.src.{args.subparser}")
        module.main(args)


if __name__ == "__main__":
    main()
