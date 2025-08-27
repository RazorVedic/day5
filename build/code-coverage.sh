#!/bin/bash

mkdir -p tmp
export GOPRIVATE=github.com/razorpay/*
go install gotest.tools/gotestsum@latest
touch tmp/test-output.log tmp/utMetrics.csv
list=$(go list ./... | grep -v "e2e")
i=1
export START_TIME=$(date +%s)
echo "start-time: ${START_TIME}"
for pkg in $list
do
    echo "pkg name",$pkg
    gotestsum -- -coverprofile=pkg-$i.cover.out -coverpkg=./... -covermode=atomic $pkg >> tmp/test-output.log
    x=$?
    i=$((i+1))
    if [[ $x -ne 0 ]]; then
        echo "Unit tests failed"
        cat tmp/test-output.log
        exit $x
    fi
done

echo "out of for loop"
cat tmp/test-output.log
#    export END_TIME=$(date -d '+330 minutes' '+%F %T');#converts UTC to IST
#    echo "end-time: ${END_TIME}"
export PASSED=$(cat tmp/test-output.log | grep DONE | awk '{s+=$2} END {print s}')
export SKIPPED=$(cat tmp/test-output.log | grep skipped | awk '{s+=$4} END {print s}')
export FAILED=$(cat tmp/test-output.log | grep failures | awk '{s+=$5} END {print s}')
echo "Test cases Passed: ${PASSED}"
echo "Test cases Skipped: ${SKIPPED}"
echo "Test cases Failed: ${FAILED}"

export END_TIME=$(date +%s)
echo "end-time: ${END_TIME}"

echo "${START_TIME},${END_TIME},${PASSED},${FAILED},${SKIPPED}" >> tmp/utMetrics.csv
echo "mode: set" > tmp/coverage.out && cat *.cover.out | grep -v mode: | sort -r | \
awk '{if($1 != last) {print $0;last=$1}}' >> tmp/coverage.out
