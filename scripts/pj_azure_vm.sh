if [ $# -eq 0 ]
    then
        echo "Options are create | destroy | stop | start"
fi

if [ $1 == 'create' ]
    then
        az group create --name MyVMResourceGroup --location uksouth
        az vm create --resource-group myVMResourceGroup --name myVM --image OpenLogic:CentOS:8_2:latest --size Standard_B1s --admin-username azureuser --generate-ssh-keys --custom-data cloud-init.txt
        PUBLICIP=$(az vm show -d -g myVMResourceGroup -n myVM --query publicIps -o tsv);ssh -q azureuser@$PUBLICIP
fi

if [ $1 == 'destroy' ]
    then
        az group delete --name myVMResourceGroup --yes

fi

if [ $1 == 'stop' ]
    then
        az vm stop --resource-group myVMResourceGroup --name myVM
fi

if [ $1 == 'start' ]
    then
        az vm start --resource-group myVMResourceGroup --name myVM
        PUBLICIP=$(az vm show -d -g myVMResourceGroup -n myVM --query publicIps -o tsv);ssh -q azureuser@$PUBLICIP
fi