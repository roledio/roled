#!/bin/bash

# https://stackoverflow.com/questions/8653126/how-to-increment-version-number-in-a-shell-script/64390598#64390598

# increment_version 1.39.3 0 : 2.0.0
# increment_version 1.39.3 1 : 1.40.0
# increment_version 1.39.3 2 : 1.39.4

increment_version() {
    local delimiter=.
    local array=($(echo "$1" | tr $delimiter '\n'))
    array[$2]=$((array[$2] + 1))
    if [ $2 -lt 2 ]; then array[2]=0; fi
    if [ $2 -lt 1 ]; then array[1]=0; fi
    echo $(
        local IFS=$delimiter
        echo "${array[*]}"
    )
}

if [ ! -z $1 ]; then
    if [ $1 == "increment" ]; then
        increment_version $2 $3
    fi
fi
