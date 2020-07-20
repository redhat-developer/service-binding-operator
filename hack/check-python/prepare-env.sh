directories=${directories:-"test/acceptance/features"}
pass=0
fail=0

export PYTHON_VENV_DIR=${PYTHON_VENV_DIR:-venv}

function prepare_venv() {
    python3 -m venv "$PYTHON_VENV_DIR" && source "$PYTHON_VENV_DIR/bin/activate"
    for req in $(find . -name 'requirements.txt'); do
        python3 "$(which pip3)" install -q -r $req;
    done
    python3 "$(which pip3)" install -q pydocstyle pyflakes vulture radon
}
