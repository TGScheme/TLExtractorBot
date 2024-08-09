package services

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/logging"
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
		logging.Fatal(err)
	}
	_, err = s.Status()
	if err != nil {
		if !environment.Uninstall {
			logging.Fatal(s.Install())
			logging.Fatal(s.Start())
			logging.Info("Service installed and started with name:", consts.ServiceName)
		} else {
			logging.Error("Service not installed")
		}
		return
	}
	if environment.Uninstall {
		if err = s.Stop(); err != nil {
			logging.Fatal("Sudo required to stop and uninstall the service")
		}
		logging.Fatal(s.Uninstall())
		logging.Info("Service uninstalled with name:", consts.ServiceName)
		return
	}
	if !service.Interactive() {
		logging.Fatal(s.Run())
	} else {
		logging.Warn("Use \"service\" command to control the service")
	}
}
