#!/usr/bin/env zsh
#
conda create --prefix ./.env python=3.11
conda init ./.env
conda activate ./.env

python -m pip install redis
python -m pip install Faker
python -m pip install jupyter
python -m pip install pytest
python -m pip install ipython
python -m pip install numpy
python -m pip install pandas
python -m pip install "python-lsp-server[all]"
