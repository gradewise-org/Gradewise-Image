#!/usr/bin/env bash

mkdir -p /root/.ssh
ENV_PATH=/root/.ssh/environment

:> $ENV_PATH

# Preserves environment variables before running sshd
function write_env_var() {
    if [ -n "${!1}" ]; then
        echo "$1=${!1}" >> $ENV_PATH
    fi
}

write_env_var ASSIGNMENT_TITLE
write_env_var AUTHENTICATION_TOKEN
write_env_var BASIC_AUTH
write_env_var SUBMISSION_URL
write_env_var SUBMIT_RESULTS_URL
write_env_var DEVEL

/usr/bin/ssh-keygen -A

/usr/sbin/sshd -D