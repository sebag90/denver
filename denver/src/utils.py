from pathlib import Path
import tomllib

def get_config():
    ROOT = Path(__file__).parent.parent.resolve()
    return tomllib.loads(Path(f"{ROOT}/config.toml").read_text())

def get_env_base_dir():
    config = get_config()
    if "venv_dir" in config:
        return Path(config["venv_dir"])
    else:
        home_dir = Path().home()
        return Path(f"{home_dir}/.denver")