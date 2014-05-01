// The nimbus package contains a packer.Builder implementation that
// builds images for Nimbus.
package nimbus

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// The unique ID for this builder
const BuilderId = "nimbusproject.nimbus"

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Information for the source instance
	Cloud       string
	SourceImage string `mapstructure:"source_image"`
	SSHUsername string `mapstructure:"ssh_username"`
	SSHPort     int    `mapstructure:"ssh_port"`
	sshTimeout  time.Duration

	// Configuration of the resulting image
	ImageName string `mapstructure:"image_name"`

	// Nimbus Cloud Client configuration
	CloudClientPath string `mapstructure:"cloud_client_path"`
	Factory         string `mapstructure:"factory"`
	Repository      string `mapstructure:"repository"`
	FactoryIdentity string `mapstructure:"factory_identity"`
	S3Id            string `mapstructure:"s3id"`
	S3Key           string `mapstructure:"s3key"`
	CanonicalId     string `mapstructure:"canonicalid"`
	Cert            string `mapstructure:"cert"`
	Key             string `mapstructure:"key"`
	MountAs         string `mapstructure:"mount_as"`

	PackerDebug   bool   `mapstructure:"packer_debug"`
	RawSSHTimeout string `mapstructure:"ssh_timeout"`

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Defaults
	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	if b.config.RawSSHTimeout == "" {
		b.config.RawSSHTimeout = "1m"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if b.config.SourceImage == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A source_image must be specified"))
	}

	if b.config.Factory == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A factory must be specified"))
	}

	if b.config.Repository == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A repository must be specified"))
	}

	if b.config.FactoryIdentity == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A factory_identity must be specified"))
	}

	if b.config.S3Id == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("An s3id must be specified"))
	}

	if b.config.S3Key == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A s3key must be specified"))
	}

	if b.config.CanonicalId == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A canonicalid must be specified"))
	}

	if b.config.Cert == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A cert must be specified"))
	}

	if b.config.Key == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("A key must be specified"))
	}

	if b.config.SSHUsername == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("An ssh_username must be specified"))
	}

	b.config.sshTimeout, err = time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}

	if b.config.ImageName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("image_name must be specified"))
	}

	if b.config.CloudClientPath == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("The path to the cloud client installation must be specified"))
	}

	templates := map[string]*string{
		"image_name":   &b.config.ImageName,
		"source_image": &b.config.SourceImage,
		"s3id": &b.config.S3Id,
		"s3key": &b.config.S3Key,
		"canonicalid": &b.config.CanonicalId,
		"cert": &b.config.Cert,
		"key": &b.config.Key,
		"cloud_client_path": &b.config.CloudClientPath,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Printf("Config: %+v", b.config)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	cloud_client_command_path := filepath.Join(b.config.CloudClientPath, "bin", "cloud-client.sh")
	cloud_client_command, err := exec.LookPath(cloud_client_command_path)
	if err != nil {
		panic(fmt.Sprintf("failed to lookup cloud client at %s", cloud_client_command_path))
	}

	log.Printf("cloud client command path: %s", cloud_client_command_path)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("cloud_client_command", cloud_client_command)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("nimbus_%s.pem", b.config.PackerBuildName),
		},
		&stepRunSourceInstance{},
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      SSHConfig(b.config.SSHUsername),
			SSHWaitTimeout: b.config.sshTimeout,
		},
		&common.StepProvision{},
		&stepCreateImage{},
	}

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	cloud_conf_path := state.Get("cloud_conf_path").(string)
	os.Remove(cloud_conf_path)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there is no image, then just return
	if _, ok := state.GetOk("image"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &artifact{
		image: b.config.ImageName,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
