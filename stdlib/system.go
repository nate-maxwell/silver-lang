package stdlib

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"silver/object"
	"strings"
	"sync"
)

const (
	silverPathEnvironmentName = "SILVER_PATH"
	silverVersionMajor        = 0
	silverVersionMinor        = 1
	silverVersionPatch        = 0
)

var silverVersion = fmt.Sprintf("%d.%d.%d", silverVersionMajor, silverVersionMinor, silverVersionPatch)

var systemEnvironmentMu sync.Mutex

// systemDefinitions contains host and process-environment information exposed
// by import("system"). Queries return empty strings when the host cannot
// provide a value.
func systemDefinitions(null *object.Null) []definition {
	return []definition{
		{name: "ENV_SILVER_PATH", value: &object.String{Value: silverPathEnvironmentName}},
		{name: "MAJOR", value: &object.Integer{Value: silverVersionMajor}},
		{name: "MINOR", value: &object.Integer{Value: silverVersionMinor}},
		{name: "PATCH", value: &object.Integer{Value: silverVersionPatch}},
		{name: "VERSION", value: &object.String{Value: silverVersion}},
		{name: "machine", fn: systemMachine},
		{name: "node", fn: systemNode},
		{name: "processor", fn: systemProcessor},
		{name: "release", fn: systemRelease},
		{name: "system", fn: systemName},
		{name: "get_path_sep", fn: systemGetPathSeparator},
		{name: "append_path", fn: systemAppendPath(null)},
		{name: "getenv", fn: systemGetenv},
		{name: "setenv", fn: systemSetenv(null)},
		{name: "environment", fn: systemEnvironment},
	}
}

func systemGetPathSeparator(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	return &object.String{Value: string(os.PathListSeparator)}
}

func systemAppendPath(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		entry, err := requireString("append_path", 0, args[0])
		if err != nil {
			return err
		}

		systemEnvironmentMu.Lock()
		defer systemEnvironmentMu.Unlock()
		value := os.Getenv(silverPathEnvironmentName)
		if value == "" {
			value = entry
		} else {
			value += string(os.PathListSeparator) + entry
		}
		if goErr := os.Setenv(silverPathEnvironmentName, value); goErr != nil {
			return newError(object.RuntimeErrorKindValue, "could not append to %s: %s", silverPathEnvironmentName, goErr)
		}
		return null
	}
}

func systemMachine(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if architecture := strings.TrimSpace(os.Getenv("PROCESSOR_ARCHITECTURE")); architecture != "" {
			return &object.String{Value: architecture}
		}
	}
	return &object.String{Value: strings.ToUpper(runtime.GOARCH)}
}

func systemNode(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	return &object.String{Value: hostname}
}

func systemProcessor(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	return &object.String{Value: processorName()}
}

func processorName() string {
	switch runtime.GOOS {
	case "windows":
		return strings.TrimSpace(os.Getenv("PROCESSOR_IDENTIFIER"))
	case "linux", "android":
		contents, err := os.ReadFile("/proc/cpuinfo")
		if err != nil {
			return ""
		}
		var modelName, hardware, processor string
		for _, line := range strings.Split(string(contents), "\n") {
			key, value, found := strings.Cut(line, ":")
			if !found {
				continue
			}
			value = strings.TrimSpace(value)
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "model name":
				if modelName == "" {
					modelName = value
				}
			case "hardware":
				if hardware == "" {
					hardware = value
				}
			case "processor":
				if processor == "" && strings.IndexFunc(value, func(character rune) bool { return character < '0' || character > '9' }) >= 0 {
					processor = value
				}
			}
		}
		if modelName != "" {
			return modelName
		}
		if hardware != "" {
			return hardware
		}
		return processor
	case "darwin":
		if value := commandOutput("sysctl", "-n", "machdep.cpu.brand_string"); value != "" {
			return value
		}
		return commandOutput("sysctl", "-n", "hw.model")
	default:
		return strings.TrimSpace(os.Getenv("PROCESSOR_IDENTIFIER"))
	}
}

func systemRelease(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	return &object.String{Value: systemReleaseName()}
}

func systemReleaseName() string {
	switch runtime.GOOS {
	case "windows":
		release := strings.TrimSpace(os.Getenv("OS"))
		if _, suffix, found := strings.Cut(release, "_"); found {
			return suffix
		}
		return release
	case "linux", "android":
		contents, err := os.ReadFile("/proc/sys/kernel/osrelease")
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(contents))
	case "darwin", "freebsd", "netbsd", "openbsd", "dragonfly", "aix", "solaris":
		return commandOutput("uname", "-r")
	default:
		return ""
	}
}

func commandOutput(name string, args ...string) string {
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func systemName(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	names := map[string]string{
		"aix":       "AIX",
		"android":   "Android",
		"darwin":    "Darwin",
		"dragonfly": "DragonFlyBSD",
		"freebsd":   "FreeBSD",
		"illumos":   "SunOS",
		"ios":       "iOS",
		"js":        "JavaScript",
		"linux":     "Linux",
		"netbsd":    "NetBSD",
		"openbsd":   "OpenBSD",
		"plan9":     "Plan9",
		"solaris":   "SunOS",
		"wasip1":    "WASI",
		"windows":   "Windows",
	}
	return &object.String{Value: names[runtime.GOOS]}
}

func systemGetenv(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	key, err := requireString("getenv", 0, args[0])
	if err != nil {
		return err
	}
	return &object.String{Value: os.Getenv(key)}
}

func systemSetenv(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		key, err := requireString("setenv", 0, args[0])
		if err != nil {
			return err
		}
		value, err := requireString("setenv", 1, args[1])
		if err != nil {
			return err
		}
		systemEnvironmentMu.Lock()
		goErr := os.Setenv(key, value)
		systemEnvironmentMu.Unlock()
		if goErr != nil {
			return newError(object.RuntimeErrorKindValue, "could not set environment variable %q: %s", key, goErr)
		}
		return null
	}
}

func systemEnvironment(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	pairs := make(map[object.HashKey]object.MapPair)
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		// Windows may expose drive-current-directory entries such as
		// =C:=C:\path. Preserve their complete key.
		if key == "" && strings.HasPrefix(entry, "=") {
			if index := strings.Index(entry[1:], "="); index >= 0 {
				index++
				key, value = entry[:index], entry[index+1:]
			}
		}
		keyObject := &object.String{Value: key}
		pairs[keyObject.HashKey()] = object.MapPair{
			Key:   keyObject,
			Value: &object.String{Value: value},
		}
	}
	return &object.Map{Pairs: pairs}
}
