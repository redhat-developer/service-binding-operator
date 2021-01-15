#!/bin/bash

_killall(){
    which killall &> /dev/null
    if [ $? -eq 0 ]; then
        killall $1
    else
        for i in "$(ps -l | grep $1)"; do if [ -n "$i" ]; then kill $(echo "$i" | sed -e 's,\s\+,#,g' | cut -d "#" -f4); fi; done
    fi
}

# Kill SBO running locally (no matter how it was started);
_killall bin/manager
_killall manager

exit 0