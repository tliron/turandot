#!/bin/sh
# install turandot on Centos. Assumes required published packages already installed by cloud-init
echo "this has been written by turandot script pj_install.sh " + $(date) >> ~/logs.txt

cd /tmp
#Install turandot binary
command -v turandot >/dev/null 2>&1 || {
      wget -O turandot https://github.com/tliron/turandot/releases/download/v0.4.0/turandot_0.4.0_linux_amd64.rpm
      sudo rpm turandot
      command -v turandot >/dev/null 2>&1 || { echo >&2 "Could not install turandot.  Aborting."; exit 1; }
}
#Install reposure binary
command -v reposure >/dev/null 2>&1 || {
      wget -O https://github.com/tliron/reposure/releases/download/v0.1.3/reposure_0.1.3_linux_amd64.rpm
      sudo rpm install reposure
      command -v reposure >/dev/null 2>&1 || { echo >&2 "Could not install reposure.  Aborting."; exit 1; }
}
#Install puccini binary
command -v puccini-tosca >/dev/null 2>&1 || {
      wget -O puccini https://github.com/tliron/puccini/releases/download/v0.17.0/puccini_0.17.0_linux_amd64.rpm
      sudo rpm install puccini
      command -v puccini-tosca >/dev/null 2>&1 || { echo >&2 "Could not install puccini.  Aborting."; exit 1; }
}

#TODO make the above a loop over an array of JSON objects each representing the required binary
cd ~/turandot
