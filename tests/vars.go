package tests

import (
	"os"
	"path"
)

// Test Path Variables
const (
	AssetsDir     = "assets"
	DockerDir     = "docker"
	DockerFile    = "Dockerfile"
	DockerTarBall = "docker.tar"
	TestScript    = "helloWorld.sh"
	TestVarScript = "envVars.sh"
)

// Test Variables
var (
	message           = "testing message"
	basicCommand      = []string{"echo", message}
	testScriptCommand = []string{"/bin/sh", "/src/" + TestScript}
	TestScriptMessage = "HELLO WORLD"
	testVarsCommand   = []string{"/bin/sh", "/src/" + TestVarScript, "$" + testEnv}
	testEnv           = "TEST"
	testVal           = "Value42"
	testCustomImage   = "taubyte/test:test2"
	testVolume        = "volume"
)

var (
	AssetsPath        string
	VolumePath        string
	DockerDirPath     string
	DockerFilePath    string
	DockerTarBallPath string
	ScriptPath        string
	VarScriptPath     string
)

func init() {
	if wd, err := os.Getwd(); err != nil {
		panic("Getting working directory failed with: " + err.Error())
	} else {
		AssetsPath = path.Join(wd, AssetsDir)
		VolumePath = path.Join(AssetsPath, testVolume)
		DockerDirPath = path.Join(AssetsPath, DockerDir)
		DockerFilePath = path.Join(DockerDirPath, DockerFile)
		DockerTarBallPath = path.Join(AssetsPath, DockerTarBall)
		ScriptPath = path.Join(VolumePath, TestScript)
		VarScriptPath = path.Join(VolumePath, TestVarScript)
	}
}
