from pathlib import Path
import subprocess
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


def rm_tree(pth):
    for child in pth.iterdir():
        if child.is_file():
            child.unlink()
        else:
            rm_tree(child)
    pth.rmdir()


def stop_docker(name):
    denver_base_dir = get_env_base_dir()
    env_dir = Path(f"{denver_base_dir}/{name}")

    if env_dir.exists():
        subprocess.run(
            [
                "docker",
                "compose",
                "-f",
                Path(f"{env_dir}/compose"),
                "down",
            ]
        )


def remove_env(name):
    denver_base_dir = get_env_base_dir()
    env_dir = Path(f"{denver_base_dir}/{name}")

    if env_dir.exists():
        stop_docker(name)

    subprocess.run(["docker", "image", "rm", f"{name}-denver_{name}"])

    if env_dir.exists():
        rm_tree(env_dir)
