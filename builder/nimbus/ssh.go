package nimbus

import (
	gossh "code.google.com/p/gosshold/ssh"
	"fmt"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/multistep"
)

func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(config)
	hostname := state.Get("hostname").(string)
	return fmt.Sprintf("%s:%d", hostname, config.SSHPort), nil
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		privateKey := state.Get("privateKey").(string)

		keyring := new(ssh.SimpleKeychain)
		if err := keyring.AddPEMKey(privateKey); err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &gossh.ClientConfig{
			User: username,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		}, nil
	}
}
