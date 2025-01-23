package services

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/kardianos/service"
	"os"
)

func Run(runner func()) {
	if environment.Debug {
		runner()
	}
	svcConfig := &service.Config{
		Name:        consts.ServiceName,
		DisplayName: consts.ServiceDisplayName,
		Description: consts.ServiceDescription,
		Arguments: []string{
			"-C",
			environment.EnvFolder,
		},
		EnvVars: map[string]string{
			"PATH": fmt.Sprintf("/opt/sdkman/candidates/java/current/bin:%s", os.Getenv("PATH")),
		},
	}
	c := &context{
		funcRun: runner,
	}
	s, err := service.New(c, svcConfig)
	if err != nil {
		gologging.Fatal(err)
	}
	_, err = s.Status()
	if err != nil {
		if !environment.Uninstall {
			if err = s.Install(); err != nil {
				gologging.Fatal(err)
			}
			if err = s.Start(); err != nil {
				gologging.Fatal(err)
			}
			gologging.Info("Service installed and started with name:", consts.ServiceName)
		} else {
			gologging.Error("Service not installed")
		}
		return
	}
	if environment.Uninstall {
		if err = s.Stop(); err != nil {
			gologging.Fatal("Sudo required to stop and uninstall the service")
		}
		if err = s.Uninstall(); err != nil {
			gologging.Fatal(err)
		}
		gologging.Info("Service uninstalled with name:", consts.ServiceName)
		return
	}
	if !service.Interactive() {
		if err = s.Run(); err != nil {
			gologging.Fatal(err)
		}
	} else {
		gologging.Warn("Use \"service\" command to control the service")
	}
}
