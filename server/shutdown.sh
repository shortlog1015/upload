#!/usr/bin/env bash
pwd=`pwd`
target=`basename $pwd`
# kill
pid=`ps -C ${target} -o pid=`
if [ -n "$pid" ]; then
        echo "Stopping old version, PID: ${pid}"
        if [ "$1" = "-f" ]; then
            # force shutdown
            echo "Force shutdown..."
            kill -9 $(ps -C ${target} -o pid=)
        else
            kill -s 2 $(ps -C ${target} -o pid=)
        fi
        # wait for program to stop
        pid=`ps -C ${target} -o pid=`
        while [ -n "$pid" ]; do
                sleep 0.2
                pid=`ps -C ${target} -o pid=`
        done
fi
echo "Done"
