from setuptools import setup, find_packages

setup(
    name="denver",
    version="1.0.5",
    license="MIT",
    packages=find_packages(),
    include_package_data=True,
    package_data={"": ["template/*", "config.toml"]},
    entry_points={
        "console_scripts": ["denver = denver.bin.run:main"],
    },
)
