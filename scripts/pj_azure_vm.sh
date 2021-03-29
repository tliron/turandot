# uses azure command to execute an action on a virtual machine

RESOURCEGROUP=MyVMResourceGroup
LOCATION=uksouth
VMNAME=myVM

function create {
    az group create --name $RESOURCEGROUP --location $LOCATION
    az vm create --resource-group $RESOURCEGROUP --name $VMNAME --image OpenLogic:CentOS:8_2:latest --size Standard_B2s --admin-username azureuser --generate-ssh-keys --custom-data scripts/pj_cloud-init.txt
}


function destroy {
    az group delete --name $RESOURCEGROUP --yes
}

function connect {
    PUBLICIP=$(az vm show -d -g $RESOURCEGROUP -n $VMNAME --query publicIps -o tsv);ssh -q azureuser@$PUBLICIP
}

case $1 in

    create)
        create
        connect
    ;;

    destroy)
        destroy
    ;;

    stop)
        az vm stop --resource-group $RESOURCEGROUP --name $VMNAME
    ;;

    start)
        az vm start --resource-group $RESOURCEGROUP --name $VMNAME
        connect
    ;;

    rebuild)
        destroy
        create
        connect
    ;;

    connect)
        connect
    ;;

    *)
        echo "ERROR: Must supply one argument which may be create | destroy | stop | start | connect | rebuild"
    ;;
esac