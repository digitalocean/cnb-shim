package cnbshim

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/buildpacks/libcnb"
)

type Release struct {
	DefaultProcessTypes map[string]string `yaml:"default_process_types,omitempty"`
}

func WriteLaunchMetadata(appDir, layersDir, targetBuildpackDir string, log logger.Logger) error {
	release, err := ExecReleaseScript(appDir, targetBuildpackDir)
	if err != nil {
		return err
	}

	procfile, err := ReadProcfile(appDir)
	if err != nil {
		return err
	}

	processTypes := make(map[string]string)
	for name, command := range release.DefaultProcessTypes {
		processTypes[name] = command
	}

	for name, command := range procfile {
		processTypes[name] = command
	}

	processes := []libcnb.Process{}
	// Starting from CNB Platform v0.5, buildpack has to set a default process.
	// We're setting web as default process, buildpacks can override this if required.
	for name, command := range processTypes {
		processes = append(processes, libcnb.Process{
			Type:    name,
			Command: command,
			Default: name == "web",
		})
	}

	launchTOML := libcnb.LaunchTOML{
		Processes: processes,
	}

	file := filepath.Join(layersDir, "launch.toml")

	if err := WriteTomlFile(file, 0644, launchTOML); err != nil {
		return err
	}

	file = filepath.Join(layersDir, "app.toml")
	return WriteTomlFile(file, 0644, launchTOML)
}

func WriteTomlFile(filename string, perm os.FileMode, value interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(value)
}

func ExecReleaseScript(appDir, targetBuildpackDir string) (Release, error) {
	releaseScript := filepath.Join(targetBuildpackDir, "bin", "release")
	_, err := os.Stat(releaseScript)
	if !os.IsNotExist(err) {
		cmd := exec.Command(releaseScript, appDir)
		cmd.Env = os.Environ()

		out, err := cmd.Output()
		if err != nil {
			return Release{DefaultProcessTypes: make(map[string]string)}, err
		}

		release := Release{}

		return release, yaml.Unmarshal(out, &release)
	} else {
		return Release{DefaultProcessTypes: make(map[string]string)}, nil
	}

}

func ReadProcfile(appDir string) (map[string]string, error) {
	processTypes := make(map[string]string)
	procfile := filepath.Join(appDir, "Procfile")
	_, err := os.Stat(procfile)
	if !os.IsNotExist(err) {

		procfileText, err := ioutil.ReadFile(procfile)
		if err != nil {
			return processTypes, err
		}

		return processTypes, yaml.Unmarshal(procfileText, &processTypes)
	} else {
		return processTypes, nil
	}
}
