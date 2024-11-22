from pathlib import Path
from setuptools import setup, find_packages


requirements = Path("requirements.txt").read_text(encoding="utf-8").split("\n")
version = Path("denver/__version__").read_text()


setup(
    name="denver",
    version=version,
    license="MIT",
    packages=find_packages(),
    install_requires=requirements,
    include_package_data=True,
    package_data={"": ["template/*", "config.toml"]},
    entry_points={
        "console_scripts": ["denver = denver.bin.run:main"],
    },
)
