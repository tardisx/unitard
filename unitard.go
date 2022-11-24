// Package unitard provides an easy-to-use method to automatically deploy
// (and undeploy) a systemd service configuration directly from your application binary.
package unitard

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates/*.service
var fs embed.FS

type Unit struct {
	name   string
	binary string

	systemCtlPath string // path to systemctl command
	unitFilePath  string
}

type UnitOpts interface{}

// NewUnit creates a new systemd unit representation, with a particular name.
// No changes will be made to the system configuration until Deploy or Undeploy
// are called.
// NewUnit will check that the local environment is suitably configured, it will
// return an error if it is not.
func NewUnit(unitName string, unitOpts ...UnitOpts) (Unit, error) {

	ok := checkName(unitName)
	if !ok {
		return Unit{}, fmt.Errorf("sorry, name '%s' is not valid", unitName)
	}
	u := Unit{
		name:   unitName,
		binary: binaryName(),
	}

	if len(unitOpts) > 0 {
		return Unit{}, fmt.Errorf("sorry, UnitOpts are not yet supported")
	}

	err := u.setupEnvironment()
	if err != nil {
		return Unit{}, err
	}
	return u, nil
}

// UnitFilename returns the full path to the systemd unit file that will be used for
// Deploy or Undeploy.
func (u Unit) UnitFilename() string {
	return fmt.Sprintf("%s%c%s.service", u.unitFilePath, os.PathSeparator, u.name)
}

// Deploy creates/overwrites the unit file, enables and starts it running.
func (u Unit) Deploy() error {

	// create/overwrite the unit file
	unitFileName := u.UnitFilename()
	f, err := os.Create(unitFileName)
	if err != nil {
		return fmt.Errorf("could not create unit file '%s': %s", unitFileName, err)
	}
	defer f.Close()

	err = u.writeTemplate(f)
	if err != nil {
		return err
	}

	// and start it up
	err = u.enableAndStartUnit()
	if err != nil {
		return err
	}

	return nil
}

func (u Unit) writeTemplate(f io.Writer) error {
	t, err := template.New("").ParseFS(fs, "templates/*.service")
	if err != nil {
		return err
	}

	data := map[string]string{
		"description": u.name,
		"execStart":   u.binary,
	}
	err = t.ExecuteTemplate(f, "basic.service", data)
	return err
}

func (u Unit) enableAndStartUnit() error {
	err := u.runExpectZero(u.systemCtlPath, "--user", "daemon-reload")
	if err != nil {
		return err
	}
	err = u.runExpectZero(u.systemCtlPath, "--user", "enable", u.name)
	if err != nil {
		return err
	}
	err = u.runExpectZero(u.systemCtlPath, "--user", "restart", u.name)
	if err != nil {
		return err
	}
	return nil
}

// Undeploy is the opposite of deploy - it will stop the service, disable it,
// remove the service file and refresh systemd.
func (u Unit) Undeploy() error {
	err := u.runExpectZero(u.systemCtlPath, "--user", "disable", u.name)
	if err != nil {
		return err
	}
	err = u.runExpectZero(u.systemCtlPath, "--user", "stop", u.name)
	if err != nil {
		return err
	}
	err = os.Remove(u.UnitFilename())
	if err != nil {
		return err
	}
	err = u.runExpectZero(u.systemCtlPath, "--user", "daemon-reload")
	if err != nil {
		return err
	}

	return nil
}

// runExpectZero runs a command + optional arguments, returning an
// error if it cannot be run, or if it returns a non-zero exit code
func (u Unit) runExpectZero(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("could not start systemctl: %s", err)
	}

	logStringA := []string{command}
	logStringA = append(logStringA, args...)
	logString := strings.Join(logStringA, " ")

	err = cmd.Wait()

	if err != nil {
		return fmt.Errorf("problem running '%s': %s", logString, err)
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("problem running '%s': exit code non-zero: %d", logString, cmd.ProcessState.ExitCode())
	}

	return nil
}

// binaryName returns the fully-qualified path to the binary on disk
func binaryName() string {
	binary, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return binary
}

func checkName(name string) bool {
	// because it is used for the filename, we should restrict it
	return regexp.MustCompile("^[a-zA-Z0-9_]+$").Match([]byte(name))
}

// setupEnvironment ensures we have systemd installed and other things ready
func (u *Unit) setupEnvironment() error {
	// check we have systemctl
	systemCtlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return fmt.Errorf("could not find systemctl: %s", err)
	}
	u.systemCtlPath = systemCtlPath

	// check we aren't root
	uid := os.Getuid()
	if uid == 0 {
		return fmt.Errorf("cannot run as root")
	}
	if uid == -1 {
		return fmt.Errorf("cannot run on windows")
	}

	// check for the service file path
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find users home dir: %s", err)
	}
	unitFileDirectory := fmt.Sprintf("%s%c%s%c%s%c%s", userHomeDir, os.PathSeparator,
		".config", os.PathSeparator,
		"systemd", os.PathSeparator,
		"user",
	)

	err = os.MkdirAll(unitFileDirectory, 0777)
	if err != nil {
		return fmt.Errorf("cannot create the user systemd path '%s': %s", unitFileDirectory, err)
	}

	sfp, err := os.Stat(unitFileDirectory)
	if err != nil {
		return fmt.Errorf("could not find user service directory '%s': %s", unitFileDirectory, err)
	}

	if !sfp.IsDir() {
		return fmt.Errorf("'%s' - not a directory", unitFileDirectory)
	}

	u.unitFilePath = unitFileDirectory
	return nil
}
