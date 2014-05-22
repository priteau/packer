package nimbus

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"strconv"
	"text/template"
	"time"
)

type stepCreateImage struct{}

type imageNameData struct {
	CreateTime string
}

func (s *stepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	instance_id := state.Get("instance_id").(string)
	cloud_client_command := state.Get("cloud_client_command").(string)
	cloud_conf_path := state.Get("cloud_conf_path").(string)

	// Parse the name of the image
	imageNameBuf := new(bytes.Buffer)
	tData := imageNameData{
		strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	t := template.Must(template.New("image").Parse(config.ImageName))
	t.Execute(imageNameBuf, tData)
	imageName := imageNameBuf.String()

	// Create the image
	ui.Say(fmt.Sprintf("Creating the image: %s", imageName))

	var stdout bytes.Buffer
	args := []string{"--conf", cloud_conf_path, "--save", "--handle", instance_id, "--newname", config.ImageName}
	if (config.PublicImage) {
		args = append(args, "--common")
	}
	cmd := exec.Command(cloud_client_command, args...)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error creating the image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("image", config.ImageName)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
