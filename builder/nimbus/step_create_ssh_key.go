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
	Debug        bool
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

	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Write out the key
		err = pem.Encode(f, &priv_blk)
		f.Close()
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
	}

	private_key := string(pem.EncodeToMemory(&priv_blk))
	state.Put("privateKey", private_key)

	// Write the public key in a temporary file
	tf, err := ioutil.TempFile("", "packer-nimbus-publickey")
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing public SSH key: %s", err))
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Saving public key: %s", tf.Name()))
	// Write out the key
	_, err = tf.WriteString(string(ssh.MarshalAuthorizedKey(pub)))
	tf.Close()
	if err != nil {
		state.Put("error", fmt.Errorf("Error saving public SSH key: %s", err))
		return multistep.ActionHalt
	}

	state.Put("ssh_public_key", tf.Name())
	return multistep.ActionContinue
}

// Clean up temporary SSH keys
func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {
	os.Remove(state.Get("ssh_public_key").(string))
}