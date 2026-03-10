package terraform

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/models"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"gorm.io/gorm"
)

type TFEngine struct {
	DB         *gorm.DB
	WorkingDir string
	ExecPath   string // Path to terraform binary
}

func NewTFEngine(db *gorm.DB, workingDir string) (*TFEngine, error) {
	execPath := "terraform" // Assume in PATH, otherwise provide absolute path
	
	// Create working directory if it doesn't exist
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		os.MkdirAll(workingDir, 0755)
	}

	return &TFEngine{
		DB:         db,
		WorkingDir: workingDir,
		ExecPath:   execPath,
	}, nil
}

// Provision spins up a new VM for an environment
func (t *TFEngine) Provision(ctx context.Context, env *models.Environment) error {
	tf, err := t.setupTF(ctx, env.Name)
	if err != nil {
		return err
	}

	// Dynamic variables for Proxmox provider
	vars := map[string]interface{}{
		"vm_name":     env.Name,
		"target_node": env.HostNode,
		"vm_id":       env.VMID,
	}

	var varOpts []tfexec.ApplyOption
	for k, v := range vars {
		varOpts = append(varOpts, tfexec.Var(fmt.Sprintf("%s=%v", k, v)))
	}

	log.Printf("Starting terraform apply for environment: %s", env.Name)
	err = tf.Apply(ctx, varOpts...)
	if err != nil {
		t.updateStatus(env, "FAILED")
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	// Capture output (IP Address)
	outputs, err := tf.Output(ctx)
	if err == nil {
		if ip, ok := outputs["vm_ip"]; ok {
			env.IPAddress = string(ip.Value)
		}
	}

	t.updateStatus(env, "HEALTHY")
	return nil
}

// Destroy tears down an environment
func (t *TFEngine) Destroy(ctx context.Context, env *models.Environment) error {
	tf, err := t.setupTF(ctx, env.Name)
	if err != nil {
		return err
	}

	log.Printf("Starting terraform destroy for environment: %s", env.Name)
	err = tf.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	t.updateStatus(env, "DESTROYED")
	return nil
}

func (t *TFEngine) setupTF(ctx context.Context, workspace string) (*tfexec.Terraform, error) {
	wsDir := filepath.Join(t.WorkingDir, workspace)
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		os.MkdirAll(wsDir, 0755)
		// Here we would copy the base .tf templates to the workspace directory
	}

	tf, err := tfexec.NewTerraform(wsDir, t.ExecPath)
	if err != nil {
		return nil, err
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	return tf, err
}

func (t *TFEngine) updateStatus(env *models.Environment, status string) {
	env.Status = status
	t.DB.Save(env)
}
