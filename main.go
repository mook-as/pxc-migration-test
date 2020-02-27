package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

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
	fmt.Printf("Total:    %10vK\n", totalKBytes)
	fmt.Printf("Empty:    %10vK\n", emptyDBSizeKBytes)
	fmt.Printf("Fraction: %10v%%\n", 100*(usedKBytes-emptyDBSizeKBytes)/totalKBytes)

	if 100*(usedKBytes-emptyDBSizeKBytes)/totalKBytes > 45 {
		return errors.New("Cannot continue, insufficient disk space to complete migration")
	}
	return nil
}

func showMounts(systemInfoGatherer Sigar) error {
	buf, err := ioutil.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return err
	}

	var mounts []string

	for _, line := range strings.Split(string(buf), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		mounts = append(mounts, fields[4])
	}

	sort.Strings(mounts)

	writer := tabwriter.NewWriter(os.Stdout, 10, 10, 1, ' ', 0)
	defer writer.Flush()

	for _, path := range mounts {
		info, err := systemInfoGatherer.GetFileSystemUsage(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", path, info)
			continue
		}
		fmt.Fprintf(writer, "%d\t%d\t%s\n", info.Used, info.Total, path)
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

	err = showMounts(&concreteSigar)
	if err != nil {
		fmt.Printf("error showing mounts: %v", err)
	}
}
