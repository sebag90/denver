from .utils import get_env_base_dir


def main(args):
    denver_base_dir = get_env_base_dir()
    if not denver_base_dir.exists():
        return 0
    
    for env_dir in denver_base_dir.iterdir():
        if env_dir.is_dir():
            print(f"* {env_dir.stem}")

    return 0
