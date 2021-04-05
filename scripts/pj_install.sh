#!/bin/sh
# install turandot on Centos. Assumes required published packages already installed by cloud-init
echo "this has been written by turandot script pj_install.sh " + $(date) >> ~/pj_turandot_install_logs.txt

cd /tmp
#Install turandot binary
wget -O turandot.rpm https://github.com/tliron/turandot/releases/download/v0.4.0/turandot_0.4.0_linux_amd64.rpm
sudo rpm -ivh turandot.rpm

wget -O reposure.rpm https://github.com/tliron/reposure/releases/download/v0.1.3/reposure_0.1.3_linux_amd64.rpm
sudo rpm -ivh reposure.rpm

#Install puccini binary
wget -O puccini.rpm https://github.com/tliron/puccini/releases/download/v0.17.0/puccini_0.17.0_linux_amd64.rpm
sudo rpm -ivh puccini.rpm

#Install kubectl and minikube here as can't get them from cloud-init packages
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.x86_64.rpm
sudo rpm -ivh minikube-latest.x86_64.rpm

#TODO make the above a loop over an array of JSON objects each representing the required binary

sudo wall -n "Completed turandot tools installation. Start a new session to use new permissions and cd to /opt/turandot"


