
directories=${directories:-"./test/acceptance/features"}
pass=0
fail=0

function prepare_venv() {
    python3 -m venv venv && source venv/bin/activate
    for req in $(find . -name 'requirements.txt'); do
        python3 "$(which pip3)" install -q -r $req;
    done
    python3 "$(which pip3)" install -q pydocstyle pyflakes vulture radon
}
