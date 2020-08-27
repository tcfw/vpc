package hyper

import (
	"bufio"
	"os"
	"strings"
	"testing"

	hyperAPI "github.com/tcfw/vpc/pkg/api/v1/hyper"
)

func TestCreateCloudInitISO(t *testing.T) {
	vm := &hyperAPI.VM{
		Id: "abcdef-test",
		UserData: []byte(`
#cloud-init
password: test
chpasswd: { expire: False }
ssh_pwauth: True
	`),
	}

	isoFile, err := createCloudInitISO(vm)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		isoFile.Close()
		os.Remove(isoFile.Name())
	}()

	t.Logf("ISO filename: %s", isoFile.Name())

	//TODO(tcfw): find some way to validate ISO file properly

	scanner := bufio.NewScanner(isoFile)
	loc := false
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "#cloud-init") {
			loc = true
			break
		}
	}

	if loc == false {
		t.Fatal("failed to find cloud-init user-data header")
	}
}
