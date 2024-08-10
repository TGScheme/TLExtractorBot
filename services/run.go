package services

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"github.com/Laky-64/gologging"
	"github.com/kardianos/service"
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
			consts.EnvFolder,
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
			gologging.Fatal(s.Install())
			gologging.Fatal(s.Start())
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
		gologging.Fatal(s.Uninstall())
		gologging.Info("Service uninstalled with name:", consts.ServiceName)
		return
	}
	if !service.Interactive() {
		gologging.Fatal(s.Run())
	} else {
		gologging.Warn("Use \"service\" command to control the service")
	}
}
