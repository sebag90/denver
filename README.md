<center>

<img alt="micro logo" src="./assets/logo.svg" width="75%"/>
</center>

Denver is a Python-based tool that simplifies the management of virtual environments using Docker. It provides an intuitive interface to create, manage, and switch between isolated development environments without cluttering your local machine.

## Features
* Easy creation of isolated development environments using Docker
* Seamless switching between different virtual environments
* Automatic dependency management
* Project-specific environment configurations
* Environment sharing and reproducibility (coming soon)


## Installation
```bash
$ git clone https://github.com/sebag90/denver.git
$ cd denver
$ pip install .
```

## Uninstall
```bash
$ pip uninstall denver
$ rm -rf ~/.denver
```

## Quick Start

1. Create a new virtual environment:
```bash
denver create <myproject>
```

2. Activate an environment:
```bash
denver activate <myproject>
```

4. List all environments:
```bash
denver list
```

5. Stop an environment:
```bash
denver stop <myproject>
```

6. Remove an environments:
```bash
denver stop <myproject>
```

7. modify your environment:
```bash
denver config -e <myproject> {requirements, env, dockerfile, compose}
```

8. rebuild your environment after a modification:
```bash
denver rebuild <myproject>
```

# Copyright
`assets/logo.svg` & `assets/mark.svg` Copyright © 2024 [Matteo Pranzetti](https://www.instagram.com/matteoprnzt/profilecard/?igsh=MXZud3AzcWNxYnRyZQ==). All Rights Reserved.
