package agent

import (
	"time"

	"code.cloudfoundry.org/clock"

	boshalert "github.com/cloudfoundry/bosh-agent/agent/alert"
	boshas "github.com/cloudfoundry/bosh-agent/agent/applier/applyspec"
	boshhandler "github.com/cloudfoundry/bosh-agent/handler"
	boshjobsuper "github.com/cloudfoundry/bosh-agent/jobsupervisor"
	boshplatform "github.com/cloudfoundry/bosh-agent/platform"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

const (
	agentLogTag = "agent"
)

//go:generate counterfeiter . CanRebooter

type CanRebooter interface {
	CanReboot() (bool, error)
}

type Agent struct {
	logger            boshlog.Logger
	mbusHandler       boshhandler.Handler
	platform          boshplatform.Platform
	actionDispatcher  ActionDispatcher
	heartbeatInterval time.Duration
	jobSupervisor     boshjobsuper.JobSupervisor
	specService       boshas.V1Service
	settingsService   boshsettings.Service
	uuidGenerator     boshuuid.Generator
	timeService       clock.Clock
	canRebooter       CanRebooter
}

func New(
	logger boshlog.Logger,
	mbusHandler boshhandler.Handler,
	platform boshplatform.Platform,
	actionDispatcher ActionDispatcher,
	jobSupervisor boshjobsuper.JobSupervisor,
	specService boshas.V1Service,
	heartbeatInterval time.Duration,
	settingsService boshsettings.Service,
	uuidGenerator boshuuid.Generator,
	timeService clock.Clock,
	canRebooter CanRebooter,
) Agent {
	return Agent{
		logger:            logger,
		mbusHandler:       mbusHandler,
		platform:          platform,
		actionDispatcher:  actionDispatcher,
		heartbeatInterval: heartbeatInterval,
		jobSupervisor:     jobSupervisor,
		specService:       specService,
		settingsService:   settingsService,
		uuidGenerator:     uuidGenerator,
		timeService:       timeService,
		canRebooter:       canRebooter,
	}
}

func (a Agent) Run() error {
	bootable, err := a.canRebooter.CanReboot()
	if err != nil {
		return bosherr.WrapError(err, "Failed to check if agent can be rebooted")
	}
	if !bootable {
		return bosherr.Error("Refusing to boot")
	}

	errCh := make(chan error, 1)

	a.actionDispatcher.ResumePreviouslyDispatchedTasks()

	go a.subscribeActionDispatcher(errCh)

	go a.generateHeartbeats(errCh)

	go func() {
		err := a.jobSupervisor.MonitorJobFailures(a.handleJobFailure(errCh))
		if err != nil {
			errCh <- err
		}
	}()

	return <-errCh
}

func (a Agent) subscribeActionDispatcher(errCh chan error) {
	defer a.logger.HandlePanic("Agent Message Bus Handler")

	err := a.mbusHandler.Run(a.actionDispatcher.Dispatch)
	if err != nil {
		err = bosherr.WrapError(err, "Message Bus Handler")
	}

	errCh <- err
}

func (a Agent) generateHeartbeats(errCh chan error) {
	a.logger.Debug(agentLogTag, "Generating heartbeat")
	defer a.logger.HandlePanic("Agent Generate Heartbeats")

	// Send initial heartbeat
	a.sendAndRecordHeartbeat(errCh)

	tickChan := time.Tick(a.heartbeatInterval)

	for {
		select {
		case <-tickChan:
			a.sendAndRecordHeartbeat(errCh)
		}
	}
}

func (a Agent) sendAndRecordHeartbeat(errCh chan error) {
	status := a.jobSupervisor.Status()
	heartbeat, err := a.getHeartbeat(status)
	if err != nil {
		err = bosherr.WrapError(err, "Building heartbeat")
		errCh <- err
		return
	}
	a.jobSupervisor.HealthRecorder(status)

	err = a.mbusHandler.Send(boshhandler.HealthMonitor, boshhandler.Heartbeat, heartbeat)
	if err != nil {
		err = bosherr.WrapError(err, "Sending heartbeat")
		errCh <- err
	}
}

func (a Agent) getHeartbeat(status string) (Heartbeat, error) {
	a.logger.Debug(agentLogTag, "Building heartbeat")
	vitalsService := a.platform.GetVitalsService()

	vitals, err := vitalsService.Get()
	if err != nil {
		return Heartbeat{}, bosherr.WrapError(err, "Getting job vitals")
	}

	spec, err := a.specService.Get()
	if err != nil {
		return Heartbeat{}, bosherr.WrapError(err, "Getting job spec")
	}

	hb := Heartbeat{
		Deployment: spec.Deployment,
		Job:        spec.JobSpec.Name,
		Index:      spec.Index,
		JobState:   status,
		Vitals:     vitals,
		NodeID:     spec.NodeID,
	}

	return hb, nil
}

func (a Agent) handleJobFailure(errCh chan error) boshjobsuper.JobFailureHandler {
	return func(monitAlert boshalert.MonitAlert) error {
		alertAdapter := boshalert.NewMonitAdapter(monitAlert, a.settingsService, a.timeService)
		if alertAdapter.IsIgnorable() {
			a.logger.Debug(agentLogTag, "Ignored monit event: ", monitAlert.Event)
			return nil
		}

		severity, found := alertAdapter.Severity()
		if !found {
			a.logger.Error(agentLogTag, "Unknown monit event name `%s', using default severity %d", monitAlert.Event, severity)
		}

		alert, err := alertAdapter.Alert()
		if err != nil {
			errCh <- bosherr.WrapError(err, "Adapting monit alert")
		}

		err = a.mbusHandler.Send(boshhandler.HealthMonitor, boshhandler.Alert, alert)
		if err != nil {
			errCh <- bosherr.WrapError(err, "Sending monit alert")
		}

		return nil
	}
}
