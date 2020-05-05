package controller

const (
	NamePrefix                = "turandot"
	PartOf                    = "turandot"
	ManagedBy                 = "turandot"
	OperatorImageName         = "tliron/turandot-operator"
	InventoryImageName        = "library/registry"
	InventorySpoolerImageName = "tliron/kubernetes-registry-spooler"
	CacheDirectory            = "/cache"

	SpoolerAppName       = "turandot-inventory"
	SpoolerContainerName = "spooler"
	SpoolDirectory       = "/spool"
)
