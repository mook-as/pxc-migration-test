package main

import (
	"errors"
	"fmt"
	"os"

	sigar "github.com/cloudfoundry/gosigar"
)

type Sigar interface {
	GetFileSystemUsage(string) (sigar.FileSystemUsage, error)
}

func RoomToMigrate(systemInfoGatherer Sigar, path string) error {

	fileSystemUsage, err := systemInfoGatherer.GetFileSystemUsage(path)
	if err != nil {
		return err
	}

	emptyDBSizeKBytes := uint64(2500000) // Approximate size of an empty Percona installation
	usedKBytes := fileSystemUsage.Used
	totalKBytes := fileSystemUsage.Total

	fmt.Printf("Sigar results:\n")
	fmt.Printf("Used:     %10vK\n", usedKBytes)
	fmt.Printf("Total:    %10vK\n", usedKBytes)
	fmt.Printf("Empty:    %10vK\n", usedKBytes)
	fmt.Printf("Fraction: %10v%%\n", 100*(usedKBytes-emptyDBSizeKBytes)/totalKBytes)

	if 100*(usedKBytes-emptyDBSizeKBytes)/totalKBytes > 45 {
		return errors.New("Cannot continue, insufficient disk space to complete migration")
	}
	return nil
}

func main() {
	path, ok := os.LookupEnv("STORE_PATH")
	if !ok {
		path = "/var/vcap/store"
	}
	concreteSigar := sigar.ConcreteSigar{}
	err := RoomToMigrate(&concreteSigar, path)
	fmt.Printf("result: %v\n", err)
}
