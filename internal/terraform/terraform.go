package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"opspilot/internal/models"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"gorm.io/gorm"
)

// TerraformClient abstracts the necessary tfexec methods for mocking
type TerraformClient interface {
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
}

// TFClientFactory defines a function type for creating TerraformClients
type TFClientFactory func(workingDir, execPath string) (TerraformClient, error)

type TFEngine struct {
	DB               *gorm.DB
	WorkingDir       string
	BaseTemplatesDir string
	ExecPath         string // Path to terraform binary
	ClientFactory    TFClientFactory
}

// defaultClientFactory is the real implementation using tfexec
func defaultClientFactory(workingDir, execPath string) (TerraformClient, error) {
	return tfexec.NewTerraform(workingDir, execPath)
}

func NewTFEngine(db *gorm.DB, workingDir string) (*TFEngine, error) {
	execPath := "terraform" // Assume in PATH, otherwise provide absolute path

	// Create working directory if it doesn't exist
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		os.MkdirAll(workingDir, 0755)
	}

	return &TFEngine{
		DB:               db,
		WorkingDir:       workingDir,
		BaseTemplatesDir: filepath.Join("terraform", "base"), // Default path
		ExecPath:         execPath,
		ClientFactory:    defaultClientFactory,
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
		"vm_name":             env.Name,
		"target_node":         env.HostNode,
		"vm_id":               env.VMID,
		"pm_api_token_id":     os.Getenv("PM_API_TOKEN_ID"),
		"pm_api_token_secret": os.Getenv("PM_API_TOKEN_SECRET"),
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
		if ipMeta, ok := outputs["vm_ip"]; ok {
			var ip string
			if err := json.Unmarshal(ipMeta.Value, &ip); err == nil {
				env.IPAddress = ip
			} else {
				// Fallback to string if unmarshal fails (e.g. if it wasn't a JSON string)
				env.IPAddress = string(ipMeta.Value)
			}
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

func (t *TFEngine) setupTF(ctx context.Context, workspace string) (TerraformClient, error) {
	wsDir := filepath.Join(t.WorkingDir, workspace)
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		os.MkdirAll(wsDir, 0755)

		// Mirror base templates
		files, err := os.ReadDir(t.BaseTemplatesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read base templates: %w", err)
		}

		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".tf" {
				src := filepath.Join(t.BaseTemplatesDir, file.Name())
				dst := filepath.Join(wsDir, file.Name())

				input, err := os.ReadFile(src)
				if err != nil {
					return nil, fmt.Errorf("failed to read template %s: %w", file.Name(), err)
				}

				if err := os.WriteFile(dst, input, 0644); err != nil {
					return nil, fmt.Errorf("failed to write template %s: %w", file.Name(), err)
				}
				log.Printf("Mirrored template: %s", file.Name())
			}
		}
	}

	tf, err := t.ClientFactory(wsDir, t.ExecPath)
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
