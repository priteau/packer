package nimbus

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
)

// StepCreateSSHKey represents a Packer build step that generates SSH key pairs.
type StepCreateSSHKey struct {
	DebugKeyPath string
}

// Run executes the Packer build step that generates SSH key pairs.
func (s *StepCreateSSHKey) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary SSH key for instance...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err := fmt.Errorf("Error creating temporary ssh key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(priv),
	}

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		err := fmt.Errorf("Error creating temporary ssh key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	tfpriv, err := ioutil.TempFile("", "packer-nimbus-privatekey")
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing private SSH key: %s", err))
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Saving private key: %s", tfpriv.Name()))
	// Write out the key
	err = pem.Encode(tfpriv, &priv_blk)
	tfpriv.Close()
	if err != nil {
		state.Put("error", fmt.Errorf("Error saving private SSH key: %s", err))
		return multistep.ActionHalt
	}

	tfpub, err := ioutil.TempFile("", "packer-nimbus-publickey")
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing public SSH key: %s", err))
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Saving public key: %s", tfpub.Name()))
	// Write out the key
	_, err = tfpub.WriteString(string(ssh.MarshalAuthorizedKey(pub)))
	tfpub.Close()
	if err != nil {
		state.Put("error", fmt.Errorf("Error saving public SSH key: %s", err))
		return multistep.ActionHalt
	}

	state.Put("ssh_private_key", tfpriv.Name())
	state.Put("ssh_public_key", tfpub.Name())
	return multistep.ActionContinue
}

// Clean up temporary SSH keys
func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {
	os.Remove(state.Get("ssh_private_key").(string))
	os.Remove(state.Get("ssh_public_key").(string))
}
