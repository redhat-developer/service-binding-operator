# How to execute the script

## Install python dependencies
pip install -r test/acceptance/features/requirements.txt

## Use this make target to check the test/acceptance python behave framework is a clean code
make lint-python-code

## Use this command to run the tests
make test-acceptance