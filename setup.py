from setuptools import setup, find_packages

setup(
    name="denver",
    version="0.6.0",
    license="MIT",
    packages=find_packages(),
    install_requires=["pick"],
    include_package_data=True,
    package_data={"": ["template/*", "config.toml"]},
    entry_points={
        "console_scripts": ["denver = denver.bin.run:main"],
    },
)
