package nimbus

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/multistep"
)

func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(config)
	hostname := state.Get("hostname").(string)
	return fmt.Sprintf("%s:%d", hostname, config.SSHPort), nil
}

func sshConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get("config").(config)
	privateKey := state.Get("privateKey").(string)

	keyring := new(ssh.SimpleKeychain)
	if err := keyring.AddPEMKey(privateKey); err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return &gossh.ClientConfig{
		User: config.SSHUsername,
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthKeyring(keyring),
		},
	}, nil
}
