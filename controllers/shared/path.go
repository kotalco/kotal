package shared

import "fmt"

const (
	// BlockchainDataSubDir is the blockchain data sub directory
	BlockchainDataSubDir = "kotal-data"
	// SecretsSubDir is the secrets (private keys, password ... etc) sub directory
	SecretsSubDir = ".kotal-secrets"
	// ConfigSubDir is the configuration sub directory
	ConfigSubDir = "kotal-config"
)

// PathData returns blockchain data directory
func PathData(homeDir string) string {
	return fmt.Sprintf("%s/%s", homeDir, BlockchainDataSubDir)
}

// PathSecrets returns secrets directory
func PathSecrets(homeDir string) string {
	return fmt.Sprintf("%s/%s", homeDir, SecretsSubDir)
}

// PathConfig returns configuration directory
func PathConfig(homeDir string) string {
	return fmt.Sprintf("%s/%s", homeDir, ConfigSubDir)
}
