#!/bin/bash

. ./hack/check-python/prepare-env.sh

# run the pyflakes for all files that are provided in $1
function check_files() {
    for source in $1
    do
        echo "$source"
        $PYTHON_VENV_DIR/bin/pyflakes "$source"
        if [ $? -eq 0 ]
        then
            echo "    Pass"
            let "pass++"
        elif [ $? -eq 2 ]
        then
            echo "    Illegal usage (should not happen)"
            exit 2
        else
            echo "    Fail"
            let "fail++"
        fi
    done
}


echo "----------------------------------------------------"
echo "Checking source files for common errors in following"
echo "directories:"
echo "$directories"
echo "----------------------------------------------------"
echo

[ "$NOVENV" == "1" ] || prepare_venv || exit 1

# checks for the whole directories
for directory in $directories
do
    pwd
    files=$(find "$directory" -path "$PYTHON_VENV_DIR" -prune -o -name '*.py' -print)

    check_files "$files"
done


if [ $fail -eq 0 ]
then
    echo "All checks passed for $pass source files"
else
    let total=$pass+$fail
    echo "$fail source files out of $total files needs to be checked and fixed"
    exit 1
fi
