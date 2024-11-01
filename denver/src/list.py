from .utils import get_env_base_dir


def main(args):
    denver_base_config = get_env_base_dir()
    for env_dir in denver_base_config.iterdir():
        if env_dir.is_dir():
            print(f"* {env_dir.stem}")
