#!/bin/bash -e

echo -e "\nChecking for presence of conflict notes in source files..."

check_results(){
    found=0
    for f in "$@"; do
        grep -PiIrnH '<<<<''<<<|>>>''>>>>|===''====' "$f"
        if [ $? -ne 1 ]; then
            found=1
        fi
    done
    if [ $found -eq 1 ]; then
        echo FAIL
    else
        echo PASS
    fi
    return $found
}

overall_failed=0

echo -e "\nChecking staged files... "
check_results $(git diff --cached --name-only | grep -v "vendor/") || overall_failed=1

echo -e "\nChecking tracked files... "
check_results $(git ls-tree --full-tree -r HEAD --name-only | grep -v "vendor/") || overall_failed=1

if [ $overall_failed -eq 0 ]; then
    echo -e "\nNone of the tracked or staged files contain strings indicating a git conflict... PASS\n"
else
    echo -e "\nThe above listed files contains strings indicating a git conflict... FAIL\n"
fi
exit $overall_failed
