# denv

denv create name --version 3.12
    creates a new directory with a dockerfile and a docker-compose file and .env

denv activate name
    runs: docker compose -f /home/seba/code/envs/base.yml up -d
    runs: docker exec -it name bash

denv deactivate name
    runs: docker compose -f /home/seba/code/envs/base.yml up -ds
