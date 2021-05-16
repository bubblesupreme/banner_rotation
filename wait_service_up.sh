#!/bin/bash
function healthCheck {
  curl http://localhost:8088 > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo 0
  else
    echo 1
  fi
}
for (( retry = 1; retry < 45; retry++ ))
do
  alive=$( healthCheck )
  if [[ $alive -eq 1 ]]; then break; fi
  sleep 1s
done

if [[ $alive -eq 1 ]]; then exit 0; fi
exit 1

