#!/bin/bash
# while read line
# do
# 	curl http://localhost:8000/v1/config/$line 2>/dev/null
# done < "${1:-/dev/stdin}"

watch curl http://localhost:8000/v1/config/k[1-10] 2>/dev/null