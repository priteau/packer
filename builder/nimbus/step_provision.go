package nimbus

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepProvision struct{}

func (*stepProvision) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)

	log.Println("Running the provision hook")
	if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*stepProvision) Cleanup(state multistep.StateBag) {}
