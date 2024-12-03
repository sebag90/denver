from pathlib import Path
from setuptools import setup, find_packages


requirements = list()
with Path("requirements.txt").open(encoding="utf-8") as infile:
    for line in infile:
        requirements.append(line.strip())

version = Path("denver/__version__").read_text(encoding="utf-8").strip()


setup(
    name="denver",
    version=version,
    license="MIT",
    packages=find_packages(),
    install_requires=requirements,
    include_package_data=True,
    package_data={"": ["template/*"]},
    entry_points={
        "console_scripts": ["denver = denver.bin.run:main"],
    },
)
