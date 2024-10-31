import importlib

from .cli import get_args


def main():
    args = get_args()

    if args.subparser is not None:
        module = importlib.import_module(f"denv.src.{args.subparser}")
        module.main(args)


if __name__ == "__main__":
    main()
