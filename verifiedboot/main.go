package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"syscall"

	"github.com/insomniacslk/systemboot/pkg/booter"
	"github.com/insomniacslk/systemboot/pkg/crypto"
	"github.com/insomniacslk/systemboot/pkg/tpm"
)

const (
	// Version of verified booter
	Version = `0.1`
	// LinuxPcrIndex for Linux measurements
	LinuxPcrIndex = 7
	// LinuxDevUUIDPath sysfs path
	LinuxDevUUIDPath = "/dev/disk/by-uuid/"
	// BaseMountPoint is the basic mountpoint Path
	BaseMountPoint = "/mnt/"
	// SignatureFileExt is the signature file extension of the FIT image
	SignatureFileExt = ".sig"
	// SignaturePublicKeyPath is the public key path for signature verifcation
	SignaturePublicKeyPath = "/etc/security/public_key.pem"
)

var banner = `

[0m[0m ________________________________________________________ [0m
[0m/ I simply cannot let such a crime against fabulosity go[0m \[0m
[0m\ uncorrected! Verified booter v` + Version + `[0m                      /[0m
[0m -------------------------------------------------------- [0m[00m
     [0m\[0m                                               [00m
      [0m\[0m                                              [00m
       [0m\[0m [38;5;188m▄▄[48;5;54;38;5;97m▄▄[49;38;5;54m▄[48;5;54m█[38;5;97m▄▄▄[49;38;5;54m▄▄[39m                                 [00m
       [38;5;54m▄[48;5;54m█[48;5;188m▄[48;5;255;38;5;254m▄[48;5;188;38;5;255m▄[48;5;97;38;5;188m▄[38;5;97m█[48;5;54m▄[38;5;54m█[48;5;97;38;5;97m████[48;5;188;38;5;188m█[38;5;255m▄▄[49;38;5;188m▄[39m                             [00m
     [38;5;54m▄[48;5;54;38;5;97m▄[48;5;97m███[48;5;54m▄[48;5;254;38;5;254m█[48;5;255;38;5;255m█[48;5;188m▄[38;5;188m█[48;5;54;38;5;255m▄▄▄▄▄[48;5;188;38;5;188m█[48;5;255;38;5;255m█[38;5;254m▄[48;5;188;38;5;255m▄[49;38;5;188m▄[39m                            [00m
    [38;5;54m▄[48;5;97m▄▄[48;5;54m████[38;5;188m▄[48;5;254;38;5;255m▄▄[48;5;255m████████[48;5;188;38;5;188m█[48;5;255;38;5;255m█[48;5;188;38;5;188m█[49;39m                            [00m
    [48;5;54;38;5;54m████[38;5;97m▄[38;5;54m█[48;5;97;38;5;188m▄[48;5;188;38;5;255m▄[48;5;255m██[38;5;117m▄[48;5;117;38;5;16m▄▄▄[48;5;188m▄[48;5;255m▄[38;5;188m▄[38;5;255m███[48;5;188;38;5;97m▄[49;38;5;54m▄[39m                           [00m
      [38;5;54m▀[48;5;97m▄[38;5;97m█[38;5;133m▄[48;5;188;38;5;188m█[48;5;255;38;5;255m██[38;5;117m▄[48;5;16;38;5;16m█[48;5;68;38;5;231m▄[38;5;68m█[48;5;231;38;5;231m██[48;5;188;38;5;16m▄[48;5;255m▄[38;5;255m██[48;5;188;38;5;188m█[48;5;97;38;5;97m█[48;5;54;38;5;54m█[49;39m                           [00m
     [38;5;54m▄▄[48;5;54;38;5;97m▄[38;5;54m█[48;5;133;38;5;188m▄[48;5;188m█[48;5;255;38;5;255m██[48;5;16;38;5;16m█[38;5;231m▄[38;5;16m█[48;5;68;38;5;68m█[48;5;231;38;5;231m██[48;5;188;38;5;16m▄[48;5;255;38;5;255m███[48;5;188;38;5;188m█[48;5;97;38;5;97m█[48;5;54;38;5;54m██[49;39m       [38;5;54m▄▄[48;5;54;38;5;133m▄▄▄▄[38;5;97m▄▄[49;38;5;54m▄▄[39m         [00m
     [48;5;54;38;5;54m█[48;5;133;38;5;133m█[48;5;54;38;5;97m▄[38;5;54m█[48;5;188;38;5;188m█[48;5;255;38;5;255m█[38;5;254m▄[38;5;255m█[48;5;16m▄▄[48;5;68m▄[48;5;231m▄▄[48;5;188m▄[48;5;255m███[48;5;188;38;5;188m█[48;5;97;38;5;97m█[38;5;54m▄[48;5;54;38;5;133m▄[38;5;54m█[49;39m     [38;5;54m▄[48;5;54;38;5;133m▄[48;5;133;38;5;97m▄[48;5;97;38;5;133m▄▄▄[48;5;54m▄[38;5;97m▄▄▄[38;5;54m█[48;5;97m▄[48;5;54;38;5;97m▄[49;38;5;54m▄[39m       [00m
      [38;5;54m▀▀▀▀[38;5;188m▀[48;5;188m█[48;5;255m▄▄▄[38;5;255m███████[48;5;188;38;5;54m▄[48;5;97;38;5;97m█[48;5;54;38;5;54m█[48;5;133;38;5;133m█[48;5;97;38;5;97m█[48;5;54;38;5;54m█[49;39m   [38;5;54m▄[48;5;54;38;5;133m▄[48;5;97;38;5;54m▄[48;5;54;38;5;133m▄▄▄▄[38;5;97m▄[48;5;97;38;5;54m▄▄[48;5;54m███[48;5;97;38;5;97m█[48;5;54m▄[49;38;5;54m▄[39m      [00m
               [38;5;188m▀▀[48;5;188m█[48;5;255;38;5;255m███[38;5;54m▄[48;5;54m█[38;5;97m▄[48;5;133;38;5;54m▄[48;5;97m▄[48;5;133;38;5;133m█[48;5;54;38;5;54m█[49;38;5;188m▄▄▄[48;5;54;38;5;133m▄▄[48;5;133;38;5;54m▄▄[49m▀▀[48;5;97m▄▄▄[48;5;54m████[48;5;97;38;5;97m██[48;5;54;38;5;54m█[49;39m      [00m
                 [48;5;188;38;5;188m█[48;5;255;38;5;255m█[38;5;54m▄[48;5;54m██[48;5;97m▄[48;5;54;38;5;133m▄[38;5;97m▄[38;5;54m█[48;5;133m▄[48;5;54;38;5;255m▄[48;5;255m███[38;5;117m▄[48;5;188;38;5;188m█[49;39m      [48;5;54;38;5;54m███[38;5;97m▄[48;5;97m██[48;5;54;38;5;54m██[49;39m      [00m
                 [48;5;188;38;5;188m█[48;5;255;38;5;255m█[48;5;54m▄[38;5;54m█[38;5;97m▄▄▄[48;5;97;38;5;54m▄[48;5;54;38;5;97m▄[38;5;54m█[48;5;255;38;5;255m██[38;5;75m▄[38;5;255m█[48;5;75m▄[48;5;255m█[48;5;188;38;5;188m█[49;39m    [38;5;54m▄[48;5;54;38;5;97m▄▄[48;5;97m█[38;5;54m▄▄[48;5;54;38;5;97m▄[48;5;97;38;5;54m▄[49m▀[39m      [00m
                  [48;5;188;38;5;188m█[48;5;255;38;5;255m██[48;5;54;38;5;54m█[38;5;97m▄▄[48;5;97m█[38;5;54m▄[48;5;54m█[48;5;255;38;5;188m▄[38;5;255m█[48;5;117m▄[48;5;255m█[48;5;75;38;5;117m▄[48;5;255;38;5;255m█[48;5;188;38;5;188m█[49;39m  [48;5;54;38;5;54m█[38;5;97m▄[48;5;97m██[38;5;54m▄[48;5;54;38;5;97m▄[48;5;97m██[38;5;54m▄[49m▀[39m       [00m
                   [48;5;188;38;5;250m▄[48;5;255;38;5;188m▄[38;5;255m█[48;5;54m▄▄▄[48;5;255m██[48;5;188m▄[48;5;255;38;5;188m▄[38;5;255m███[48;5;188;38;5;188m█[49;39m    [48;5;54;38;5;54m█[48;5;97m▄[48;5;54;38;5;97m▄[48;5;97m███[38;5;54m▄[48;5;54m█[38;5;97m▄[38;5;54m█[49;39m      [00m
                   [48;5;250;38;5;250m█[38;5;254m▄[48;5;188;38;5;188m█[48;5;255;38;5;255m██[48;5;188;38;5;188m█[49m▀▀[48;5;188;38;5;250m▄[38;5;254m▄[38;5;250m▄[48;5;255;38;5;188m▄[38;5;255m█[48;5;188m▄▄[49;38;5;188m▄[39m   [48;5;54;38;5;54m█[48;5;97;38;5;97m██[38;5;54m▄[48;5;54m██[38;5;97m▄[48;5;97m█[48;5;54;38;5;54m██[49;39m [48;5;54;38;5;54m██[49m▄[39m [00m
                   [48;5;250;38;5;250m█[48;5;254;38;5;254m█[48;5;188;38;5;188m█[48;5;255;38;5;255m██[48;5;188;38;5;188m█[49;39m   [48;5;250;38;5;250m█[48;5;254;38;5;254m█[48;5;188m▄[48;5;255;38;5;188m▄[38;5;255m██[48;5;188;38;5;188m█[49;39m   [38;5;54m▀▀▀[48;5;54m██[38;5;97m▄[48;5;97m█[38;5;54m▄[48;5;54;38;5;97m▄[38;5;54m█[49m▄[48;5;54m█[38;5;97m▄[48;5;97;38;5;54m▄[49m▀[39m[00m
                  [48;5;250;38;5;250m█[48;5;254;38;5;254m█[48;5;188;38;5;188m█[48;5;255;38;5;255m███[48;5;188;38;5;188m█[49;39m   [48;5;250;38;5;250m█[48;5;254;38;5;254m██[48;5;188;38;5;188m█[48;5;255;38;5;255m██[48;5;188m▄[49;38;5;188m▄[39m     [38;5;54m▀▀▀[48;5;97m▄[48;5;54;38;5;97m▄[48;5;97m█[48;5;54;38;5;54m█[49m▀[48;5;97m▄▄[49m▀[39m [00m
                 [38;5;250m▄[48;5;250;38;5;254m▄[48;5;254m█[48;5;188;38;5;188m█[48;5;255;38;5;255m███[48;5;188;38;5;188m█[49;39m   [48;5;250;38;5;250m█[48;5;254;38;5;254m██[48;5;188;38;5;188m█[48;5;255;38;5;255m███[48;5;188;38;5;188m█[49;39m         [38;5;54m▀▀▀[39m     [00m
                [48;5;250;38;5;250m█[38;5;254m▄[48;5;254m█[48;5;188;38;5;188m█[48;5;255;38;5;255m████[48;5;188;38;5;188m█[49;39m   [48;5;250;38;5;250m█[48;5;254;38;5;254m██[48;5;188;38;5;188m█[48;5;255;38;5;255m███[48;5;188m▄[49;38;5;188m▄[39m                [00m
                [38;5;250m▀▀[48;5;188;38;5;188m█[38;5;255m▄[48;5;255m███[38;5;188m▄[49m▀[39m   [38;5;250m▀▀▀[48;5;188;38;5;188m█[48;5;255;38;5;255m████[48;5;188;38;5;188m█[49;39m                [00m
                  [38;5;188m▀▀▀▀▀▀[39m       [38;5;188m▀▀▀▀▀▀[39m                [00m
                                                     [00m

`

var (
	doDebug      = flag.Bool("D", false, "Print debug output")
	bootMode     = flag.String("b", "", "Set the boot mode (verified, measured, both)")
	deviceUUID   = flag.String("d", "", "Block device identified by UUID which should be used")
	fitFilePath  = flag.String("f", "", "FIT image file path on block device")
	debug        func(string, ...interface{})
	publicKey    []byte
	tpmInterface tpm.TPM
)

func main() {
	flag.Parse()
	log.Print(banner)

	debug = func(string, ...interface{}) {}
	if *doDebug {
		debug = log.Printf
	}

	//TODO hw rng

	// Initialize the TPM
	if *bootMode == booter.BootModeMeasured || *bootMode == booter.BootModeBoth {
		tpmInterface, err := tpm.NewTPM()
		if err != nil {
			debug("Error")
			//TODO Recoverer
		}

		if err = tpmInterface.SetupTPM(); err != nil {
			debug("Error")
			//TODO Recoverer
		}
	}

	// Check if device by UUID exists
	devicePath := LinuxDevUUIDPath + *deviceUUID
	if _, err := os.Stat(devicePath); err != nil {
		debug("Error")
		//TODO Recoverer
	}

	// Check supported filesystems
	filesystems, err := diskutils.GetSupportedFilesystems()
	if err != nil {
		log.Fatal(err)
	}

	// Mount device under base path
	mountPath := path.Join(BaseMountPoint, *deviceUUID)
	mountPoint, err := diskutils.Mount(devicePath, mountPath, filesystems)
	if err != nil {
		debug("Failed to mount %s on %s: %v", devicePath, mountPath, err)
		//TODO Recoverer
	}

	// Check FIT image existence
	fitImage := mountPath + *fitFilePath
	if _, err = os.Stat(fitImage); err != nil {
		debug("Error")
		//TODO Recoverer
	}

	// Read fitImage into memory
	fitImageData, err := ioutil.ReadFile(fitImage)
	if err != nil {
		debug("Error")
		//TODO Recoverer
	}

	// Verify signature of FIT image on device
	if *bootMode == booter.BootModeVerified || *bootMode == booter.BootModeBoth {
		fitImageSignature := mountPath + *fitFilePath + SignatureFileExt
		if _, err := os.Stat(fitImageSignature); err != nil {
			debug("Error")
			//TODO Recoverer
		}

		// Read fit image signature into memory
		fitImageSignatureData, err := ioutil.ReadFile(fitImageSignature)
		if err != nil {
			debug("Error")
			//TODO Recoverer
		}

		publicKey, err := crypto.LoadPublicKeyFromFile(SignaturePublicKeyPath)
		if err != nil {
			debug("Error")
			//TODO Recoverer
		}

		if err := crypto.VerifyRsaSha256Pkcs1v15Signature(publicKey, fitImageData, fitImageSignatureData); err != nil {
			debug("Error")
			//TODO Recoverer
		}
	}

	// Measure FIT image into linux PCR
	if *bootMode == booter.BootModeMeasured || *bootMode == booter.BootModeBoth {
		tpmInterface.Measure(LinuxPcrIndex, fitImageData)
	}

	// TODO Load FIT and Kexec

	// Unmount Device
	syscall.Unmount(mountPoint.Path, syscall.MNT_DETACH)
}
