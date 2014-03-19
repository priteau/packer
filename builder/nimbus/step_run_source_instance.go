package nimbus

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type stepRunSourceInstance struct {
	instance_id string
	hostname string
}

func (s *stepRunSourceInstance) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	cloud_client_command := state.Get("cloud_client_command").(string)
	cloud_conf_path := state.Get("cloud_conf_path").(string)

	ui.Say("Launching a source Nimbus instance...")

	var stdout bytes.Buffer
	cmd := exec.Command(cloud_client_command, "--conf", cloud_conf_path, "--run", "--hours", "1", "--name", config.SourceImage)
	cmd.Stdout = &stdout

	log.Println("Executing command: ", cmd.Args)
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error launching source instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, line := range strings.Split(stdout.String(), "\n") {
		re_id:= regexp.MustCompile(`Creating workspace \"(vm-[0-9]+)\"`)
		submatch := re_id.FindStringSubmatch(line)
		if len(submatch) == 2 {
			s.instance_id = submatch[1]
		}

		re_host := regexp.MustCompile(`Hostname: (.+)$`)
		submatch_host := re_host.FindStringSubmatch(line)
		if len(submatch_host) == 2 {
			s.hostname = submatch_host[1]
		}
	}

	log.Printf("instance id: %s", s.instance_id)
	log.Printf("hostname: %s", s.hostname)

	state.Put("instance_id", s.instance_id)
	state.Put("hostname", s.hostname)

	return multistep.ActionContinue
}

func (s *stepRunSourceInstance) Cleanup(state multistep.StateBag) {
	if s.instance_id == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	if state.Get("image") != "" {
		ui.Say("Nimbus instance already terminated by image creation...")
		return
	}

	ui.Say("Terminating the source Nimbus instance...")

	cloud_client_command := state.Get("cloud_client_command").(string)
	cloud_conf_path := state.Get("cloud_conf_path").(string)

	var stdout bytes.Buffer
	cmd := exec.Command(cloud_client_command, "--conf", cloud_conf_path, "--terminate", "--handle", s.instance_id)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("Error terminating instance, may still be around: %s", err))
		return
	}
}
